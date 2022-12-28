package transform

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dop251/goja"
	"github.com/pingcap/errors"
	"github.com/siddontang/go-log/log"
	"github.com/vnvo/prensio/config"
	"github.com/vnvo/prensio/pipeline/cdc_event"
)

const (
	ACTION_CONT = iota
	ACTION_DROP
)

type Transform struct {
	rules *[]rule
}

type rule struct {
	name    string
	schemaP *regexp.Regexp
	tableP  *regexp.Regexp
	tFunc   string
	runtime *goja.Runtime
}

func NewTransform(conf *config.CDCConfig) (*Transform, error) {

	rules := make([]rule, 0)

	for _, rconf := range conf.TransformRules {
		ms, err := regexp.Compile(rconf.MatchSchema)
		if err != nil {
			return nil, err
		}

		mt, err := regexp.Compile(rconf.MatchTable)
		if err != nil {
			return nil, err
		}

		tRule := rule{
			rconf.Name,
			ms,
			mt,
			rconf.TransformFunc,
			getTransformRuntime(rconf.Name, rconf.TransformFunc),
		}

		rules = append(rules, tRule)
	}

	return &Transform{&rules}, nil
}

func getTransformRuntime(ruleName string, transformFunc string) *goja.Runtime {
	log.Infof("creating transformer runtime for: %s, %v", ruleName, transformFunc)
	if len(strings.TrimSpace(transformFunc)) == 0 {
		log.Info("no transform func is provided, skipping ...")
		return nil
	}

	vm := goja.New()

	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	vm.Set("ACTION_DROP", ACTION_DROP)
	vm.Set("ACTION_CONT", ACTION_CONT)
	registerTransformers(vm)

	goja.MustCompile("", transformFunc, false)

	_, err := vm.RunString(transformFunc)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	_, ok := goja.AssertFunction(vm.Get("transform"))
	if !ok {
		log.Errorf("no 'transform' function found in the script")
		log.Error(transformFunc)
		panic(ok)
	}

	return vm
}

func (t *Transform) Apply(e *cdc_event.CDCEvent) (int, error) {
	verdict := ACTION_CONT
	var err error

	for _, rule := range *t.rules {
		verdict, err = rule.ApplyRule(e)
		log.Debugf("verdict=%d", verdict)
		if err != nil {
			return verdict, errors.Errorf("rule=%s %s", rule.name, err)
		}
		if verdict == ACTION_DROP {
			log.Info("ACTION_DROP, dropping the event.")
			return verdict, nil
		}
	}

	return verdict, nil
}

func (tr *rule) ApplyRule(e *cdc_event.CDCEvent) (int, error) {
	log.Debugf(
		"[%s] transform. rule-name: %s, match-schema: %s, event-schema: %s, match-table: %s, event-table: %s",
		e.Meta.Pipeline,
		tr.name,
		tr.schemaP.String(),
		e.Schema,
		tr.tableP.String(),
		e.Table,
	)

	if tr.runtime == nil {
		return ACTION_CONT, nil
	}

	if tr.schemaP.MatchString(e.Schema) && tr.tableP.MatchString(e.Table) {
		log.Debug("must apply: true")

		vm := tr.runtime
		tFunc, ok := goja.AssertFunction(vm.Get("transform"))
		if !ok {
			log.Errorf("no 'transform' function found in the script")
			return 0, fmt.Errorf("'transform' function not found. rule=%s", tr.name)
		}

		verdict, err := tFunc(goja.Undefined(), vm.ToValue(e))
		if err != nil {
			log.Errorf("[%s] transform error: %s", e.Meta.Pipeline, err)
			return 0, err
		}

		//verdict = verdict.Export().(int64)
		return int(verdict.Export().(int64)), nil
	}

	log.Debug("must apply: false")
	return ACTION_CONT, nil
}
