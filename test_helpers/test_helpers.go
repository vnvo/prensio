package test_helpers

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/go-mysql-org/go-mysql/client"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	tc "github.com/testcontainers/testcontainers-go"
)

type TestContext struct {
	MysqlPort  int
	ZK1Port    int
	ZK2Port    int
	Kafka1Port int
	Kafka2Port int
	Kafka3Port int

	DBUser string
	DBPass string

	SeedPath string
	TestUUID string

	ComposePath string
	Compose     *tc.LocalDockerCompose
}

type kLogger struct{}

func (*kLogger) Printf(msg string, params ...interface{}) {
	fmt.Printf(msg, params...)
}

func NewTestContext() *TestContext {
	rand.Seed(time.Now().UnixNano())
	min := 15000
	max := 30000
	mysPort := rand.Intn(max-min+1) + min

	seedPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	uid := strings.ToLower((uuid.New()).String())

	return &TestContext{
		MysqlPort:  mysPort,
		ZK1Port:    mysPort + 1,
		ZK2Port:    mysPort + 2,
		Kafka1Port: mysPort + 3,
		Kafka2Port: mysPort + 4,
		Kafka3Port: mysPort + 5,

		DBUser: "root",
		DBPass: "root",

		SeedPath: seedPath,
		TestUUID: uid,

		ComposePath: seedPath + "/../docker-compose-testenv.yaml",
		Compose:     nil,
	}
}

func (ts *TestContext) GetDBAddr() string {
	return fmt.Sprintf("localhost:%d", ts.MysqlPort)
}

func (ts *TestContext) GetAllKafkaBrokers() []string {
	return []string{
		fmt.Sprintf("localhost:%d", ts.Kafka1Port),
		fmt.Sprintf("localhost:%d", ts.Kafka2Port),
		fmt.Sprintf("localhost:%d", ts.Kafka3Port),
	}
}

func (ts *TestContext) InsertAndReadOne(query string, kafkaTopic string) (*mysql.Result, *kafka.Message, error) {
	dbConn, err := client.Connect(
		ts.GetDBAddr(),
		ts.DBUser, ts.DBPass,
		"test_mysql_ref_db_01")

	if err != nil {
		return nil, nil, err
	}

	dbRet, err := dbConn.Execute(query)

	if err != nil {
		return nil, nil, err
	}

	kReader, _ := ts.newKafkaReader(kafkaTopic)
	defer kReader.Close()

	fmt.Println("read start", time.Now().Local())
	msg, err := kReader.ReadMessage(context.Background())
	fmt.Println("read end", time.Now().Local())

	if err != nil {
		return nil, nil, err
	}

	return dbRet, &msg, nil
}

func (ts *TestContext) newKafkaReader(topic string) (*kafka.Reader, error) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: ts.GetAllKafkaBrokers(),
		Topic:   topic,
	})

	return r, nil
}
