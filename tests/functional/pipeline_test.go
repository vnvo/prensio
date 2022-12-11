package test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vnvo/prensio/config"
	"github.com/vnvo/prensio/pipeline"
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

func TestSimplePipelineEventIsDelivered(t *testing.T) {
	table := "test_mysql_ref_db_01.test_ref_table_01"
	randomInt := rand.Intn(10) + 1
	randomText := fmt.Sprintf("test-value-%d", randomInt)
	q := fmt.Sprintf(
		"insert into %s (int_col, text_col) values (%d, '%s')",
		table, randomInt, randomText,
	)

	_, kMsg, err := testState.InsertAndReadOne(
		q,
		table,
	)

	assert.NoError(t, err, "failed to insert and receive one event", err)

	payload := map[string]interface{}{}
	err = json.Unmarshal(kMsg.Value, &payload)
	assert.NoError(t, err, "decoding message payload to json failed")

	t.Log(kMsg.Topic)
	t.Log(string(kMsg.Value))
}
