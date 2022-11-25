package test

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	//_ "github.com/go-sql-driver/mysql"
	//"github.com/jmoiron/sqlx"

	"github.com/go-mysql-org/go-mysql/client"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	DB_SRV_NAME      = "percona8-1"
	KAFKA_SRV_NAME_1 = "kafka-1"
	KAFKA_SRV_NAME_2 = "kafka-1"
	KAFKA_SRV_NAME_3 = "kafka-1"
)

var testState *TestState

type TestLogConsumer struct {
}

func (g *TestLogConsumer) Accept(l tc.Log) {
	fmt.Println(string(l.Content))
}

func setupTestEnv(ctx context.Context) {
	testState = NewTestState()
	fmt.Println("Test State: ", testState)
	composePath := []string{testState.ComposePath}
	fmt.Println(testState)

	compose := tc.NewLocalDockerCompose(composePath, testState.TestUUID)

	// add random ports for the containers external access to prevent conflicts on parallel runs
	waitService := fmt.Sprintf("kafka_3_%s", testState.TestUUID)
	waitPort := fmt.Sprintf("%d", testState.Kafka3Port)
	fmt.Println(waitPort)
	err := compose.
		WithCommand([]string{"up", "-d"}).
		WithEnv(map[string]string{
			"uid":              testState.TestUUID,
			"TEST_DB_PORT":     strconv.Itoa(testState.MysqlPort),
			"TEST_ZK1_PORT":    strconv.Itoa(testState.ZK1Port),
			"TEST_ZK2_PORT":    strconv.Itoa(testState.ZK2Port),
			"TEST_KAFKA1_PORT": strconv.Itoa(testState.Kafka1Port),
			"TEST_KAFKA2_PORT": strconv.Itoa(testState.Kafka2Port),
			"TEST_KAFKA3_PORT": strconv.Itoa(testState.Kafka3Port),
		}).
		WaitForService(waitService, wait.ForExposedPort()).
		Invoke()

	if err.Error != nil {
		compose.Down()
		panic(err.Error)
	}

	fmt.Println(compose.Services)
	testState.Compose = compose
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	setupTestEnv(ctx)
	time.Sleep(time.Second * 3)
	status := m.Run()
	testState.Compose.Down()

	os.Exit(status)
}

func TestMysqlContainerIsReachable(t *testing.T) {
	Addr := fmt.Sprintf("localhost:%d", testState.MysqlPort)
	fmt.Println(Addr)
	dbConn, err := client.Connect(
		testState.GetDBAddr(),
		testState.DBUser, testState.DBPass,
		"test_mysql_ref_db_01")

	assert.NoError(t, err)
	assert.NoError(t, dbConn.Ping())
	dbConn.Close()
}

func TestKafkaBrokersAreReachable(t *testing.T) {
	for _, addr := range testState.GetAllKafkaBrokers() {
		conn, err := kafka.Dial("tcp", addr)
		assert.NoError(t, err)
		conn.Close()
	}
}

func TestKafkaControllerBrokerIsAvailable(t *testing.T) {

	addrs := testState.GetAllKafkaBrokers()
	conn, err := kafka.Dial("tcp", addrs[1])
	assert.NoError(t, err)

	controller, err := conn.Controller()
	assert.NoError(t, err)
	fmt.Println(controller)

	topic := "test-topic"

	var controllerConn *kafka.Conn
	controllerConn, err = kafka.Dial(
		"tcp",
		net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))

	assert.NoError(t, err)
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	assert.NoError(t, err)

	partitions, err := controllerConn.ReadPartitions()
	assert.NoError(t, err)
	m := map[string]struct{}{}

	for _, p := range partitions {
		m[p.Topic] = struct{}{}
	}
	for k := range m {
		fmt.Println(k)
	}

	conn.Close()

}
