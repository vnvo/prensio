package test

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-mysql-org/go-mysql/client"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testState *TestState

type TestLogConsumer struct {
}

func (g *TestLogConsumer) Accept(l tc.Log) {
	fmt.Println(string(l.Content))
}

func setupTestEnv(ctx context.Context) {
	testState = NewTestState()
	composePath := []string{testState.ComposePath}
	fmt.Println(testState)

	compose := tc.NewLocalDockerCompose(composePath, testState.TestUUID)
	waitService := fmt.Sprintf("kafka_3_%s", testState.TestUUID)

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
		WaitForService(waitService, wait.ForLog("started (kafka.server.KafkaServer)")).
		Invoke()

	if err.Error != nil {
		compose.Down()
		panic(err.Error)
	}

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
	t.Log(testState)
	Addr := fmt.Sprintf("localhost:%d", testState.MysqlPort)
	t.Log(Addr)
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
	t.Log(controller)

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

	for _, p := range partitions {
		t.Logf("available topic: %s", p.Topic)
	}

	conn.Close()

}
