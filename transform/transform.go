package transform

import (
	"fmt"
	"regexp"

	"github.com/dop251/goja"
	"github.com/siddontang/go-log/log"
	"github.com/vnvo/go-mysql-kafka/cdc_event"
	"github.com/vnvo/go-mysql-kafka/config"
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

func getTransformRuntime(ruleName string, transform_func string) *goja.Runtime {
	log.Debugf("creating transformer runtime for: %s", ruleName)
	vm := goja.New()

	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	registerTransformers(vm)

	_, err := vm.RunString(transform_func)
	if err != nil {
		panic(err)
	}

	return vm
}

func (t *Transform) Apply(cdc_e *cdc_event.CDCEvent) error {
	for _, rule := range *t.rules {
		err := rule.Apply(cdc_e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tr *rule) Apply(cdc_e *cdc_event.CDCEvent) error {
	log.Debugf(
		"[%s] transform. rule-name: %s, match-schema: %s, event-schema: %s, match-table: %s, event-table: %s",
		cdc_e.Meta.Pipeline,
		tr.name,
		tr.schemaP.String(),
		cdc_e.Schema,
		tr.tableP.String(),
		cdc_e.Table,
	)

	if tr.schemaP.MatchString(cdc_e.Schema) && tr.tableP.MatchString(cdc_e.Table) {
		log.Debug("must apply: true")

		vm := tr.runtime
		tFunc, ok := goja.AssertFunction(vm.Get("transform"))
		if !ok {
			log.Errorf("no 'transform' function found in the script")
			return fmt.Errorf("'transform' function not found. rule=%s", tr.name)
		}

		//ignoring the return from transform script here
		_, err := tFunc(goja.Undefined(), vm.ToValue(cdc_e))
		if err != nil {
			log.Errorf("[%s] transform error: %s", cdc_e.Meta.Pipeline, err)
			return err
		}
	}

	return nil
}
