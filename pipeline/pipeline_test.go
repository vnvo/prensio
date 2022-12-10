package pipeline_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/go-mysql-org/go-mysql/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/segmentio/kafka-go"
	"github.com/vnvo/go-mysql-kafka/config"
	"github.com/vnvo/go-mysql-kafka/pipeline"
)

var _ = Describe("Pipeline", Ordered, func() {
	var p pipeline.CDCPipeline
	var cfg config.CDCConfig

	BeforeAll(func() {
		time.Sleep(time.Second)
		cfg = config.NewCDCConfig(
			testCtx.SeedPath + "/../tests/test_confs/simple_test.toml")
		cfg.Mysql.Addr = testCtx.GetDBAddr()
		cfg.KafkaSink.Addr = strings.Join(testCtx.GetAllKafkaBrokers(), ",")

		p = pipeline.NewCDCPipeline("simple-test", &cfg)
		p.Init()
		ctx := context.Background()
		go func() {
			p.Run(ctx)
		}()

		time.Sleep(time.Second * 3)

	})

	AfterAll(func() {
		p.Close()
	})

	Describe("when setup is done", func() {
		Context("and test context is created", func() {
			var dbConn *client.Conn
			var err error
			It("mysql container is reachable", func(ctx SpecContext) {
				fmt.Println(testCtx.GetDBAddr())
				dbConn, err = client.Connect(
					testCtx.GetDBAddr(),
					testCtx.DBUser, testCtx.DBPass,
					"test_mysql_ref_db_01")

				Expect(err).Should(BeNil())
				//DeferCleanup(dbConn.Close)
			}, SpecTimeout(time.Second*5))

			It("mysql responds to ping", func(ctx SpecContext) {
				Expect(dbConn.Ping()).To(Succeed())
			}, SpecTimeout(time.Millisecond*500))

			It("kafka brokers are reachable", func() {
				for _, addr := range testCtx.GetAllKafkaBrokers() {
					conn, err := kafka.Dial("tcp", addr)
					Expect(err).Should(BeNil())
					conn.Close()
				}
			})
		})

		Context("and source/sink are working", func() {
			var (
				payload    map[string]interface{}
				randomInt  int
				randomText string
			)

			It("db schemas and tables are accessible", func() {
				ret, err := p.Query(
					fmt.Sprintf("select schema_name from information_schema.schemata where schema_name = '%s' or schema_name = '%s'",
						"to_be_ignored_db",
						"test_mysql_ref_db_01",
					),
				)

				Expect(err).Should(BeNil())
				Expect(len(ret.Values)).Should(Equal(2))
			})

			It("end to end event delivery is working", func() {
				table := "test_mysql_ref_db_01.test_ref_table_01"
				randomInt = rand.Intn(10) + 1
				randomText = fmt.Sprintf("test-value-%d", randomInt)
				q := fmt.Sprintf(
					"insert into %s (int_col, text_col) values (%d, '%s')",
					table, randomInt, randomText,
				)

				GinkgoWriter.Println("injecting change ...")

				_, kMsg, err := testCtx.InsertAndReadOne(
					q,
					table,
				)

				Expect(err).Should(BeNil())
				payload = map[string]interface{}{}
				err = json.Unmarshal(kMsg.Value, &payload)
				Expect(err).Should(BeNil())
				GinkgoWriter.Println("test event payload:", payload)
			})

			It("event payload has correct captured changes", func() {
				GinkgoWriter.Println("test event payload:", payload, randomInt, randomText)
				GinkgoWriter.Println("action:", payload["action"])
				GinkgoWriter.Println("after:", payload["after"])
				GinkgoWriter.Println("after.int_col:", payload["after"].([]interface{})[0].(map[string]interface{})["int_col"])

				//json decoder return int as float64 by default
				gotIntCol := payload["after"].([]interface{})[0].(map[string]interface{})["int_col"].(float64)
				gotTextCol := payload["after"].([]interface{})[0].(map[string]interface{})["text_col"].(string)
				Expect(payload["action"]).Should(Equal("insert"))
				Expect(int(gotIntCol)).Should(Equal(randomInt))
				Expect(gotTextCol).Should(Equal(randomText))

			})
		})
	})

})
