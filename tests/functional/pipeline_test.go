package test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vnvo/go-mysql-kafka/config"
	"github.com/vnvo/go-mysql-kafka/pipeline"
)

func TestSimplePipelineSetupIsSuccessful(t *testing.T) {
	time.Sleep(time.Second * 3)
	c := config.NewCDCConfig(testState.SeedPath + "/../test_confs/simple_test.toml")
	t.Log(c)

	c.Mysql.Addr = fmt.Sprintf("localhost:%d", testState.MysqlPort)
	c.KafkaSink.Addr = strings.Join(testState.GetAllKafkaBrokers(), ",")

	p := pipeline.NewCDCPipeline("simple-test", &c)

	p.Init()
	ctx := context.Background()
	go func() {
		p.Run(ctx)
	}()

	ret, err := p.Query(
		fmt.Sprintf("select schema_name from information_schema.schemata where schema_name = '%s' or schema_name = '%s'",
			"to_be_ignored_db",
			"test_mysql_ref_db_01",
		),
	)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(ret.Values))

	for _, row := range ret.Values {
		for _, val := range row {
			t.Log(val.Type, string(val.AsString()))
		}
	}
	t.Log(ret.Values)

	//insert into db
	//read from kafka and validate

}

func TestSimplePipelineEventIsSuccessful(t *testing.T) {
	dbRet, kMsg, err := testState.InsertAndReadOne(
		"insert into test_mysql_ref_db_01.test_ref_table_01 (int_col, text_col) values (1, 'test-value')",
		"test_mysql_ref_db_01.test_ref_table_01",
	)

	assert.NoError(t, err, "failed to insert and receive on event")

	t.Log(dbRet)
	t.Log(kMsg.Topic)
	t.Log(string(kMsg.Value))

}
