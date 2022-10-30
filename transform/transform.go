package transform

import (
	"fmt"
	"regexp"

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

		t_rule := rule{
			rconf.Name,
			ms,
			mt,
		}

		rules = append(rules, t_rule)
	}

	return &Transform{&rules}, nil
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
	fmt.Printf(
		" === transform rule - name=%v, rule.schema-re=%v, event-schema=%v, rule.table-re=%v, event-table=%v\n",
		tr.name, tr.schemaP, cdc_e.Schema, tr.tableP, cdc_e.Table)

	if tr.schemaP.MatchString(cdc_e.Schema) && tr.tableP.MatchString(cdc_e.Table) {
		fmt.Println(" === must apply: true")
	}
	return nil
}
