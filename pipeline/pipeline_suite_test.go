package pipeline_test

import (
	"fmt"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	th "github.com/vnvo/go-mysql-kafka/test_helpers"
)

var testCtx *th.TestContext

func TestPipeline(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pipeline Suite")
}

var _ = BeforeSuite(func() {
	Expect(setupContext()).To(Succeed())
})

var _ = AfterSuite(func() {
	err := testCtx.Compose.Down()
	Expect(err.Error).Should(BeNil())
})

func setupContext() error {
	testCtx = th.NewTestContext()
	composePath := []string{testCtx.ComposePath}
	fmt.Println(testCtx)

	compose := tc.NewLocalDockerCompose(composePath, testCtx.TestUUID)
	waitService := fmt.Sprintf("kafka_3_%s", testCtx.TestUUID)

	err := compose.
		WithCommand([]string{"up", "-d"}).
		WithEnv(map[string]string{
			"uid":              testCtx.TestUUID,
			"TEST_DB_PORT":     strconv.Itoa(testCtx.MysqlPort),
			"TEST_ZK1_PORT":    strconv.Itoa(testCtx.ZK1Port),
			"TEST_ZK2_PORT":    strconv.Itoa(testCtx.ZK2Port),
			"TEST_KAFKA1_PORT": strconv.Itoa(testCtx.Kafka1Port),
			"TEST_KAFKA2_PORT": strconv.Itoa(testCtx.Kafka2Port),
			"TEST_KAFKA3_PORT": strconv.Itoa(testCtx.Kafka3Port),
		}).
		WaitForService(waitService, wait.ForLog("started (kafka.server.KafkaServer)")).
		Invoke()

	if err.Error != nil {
		compose.Down()
		return err.Error
	}

	testCtx.Compose = compose
	return nil
}

func cleanUpTestContext() error {
	err := testCtx.Compose.Down()
	if err.Error != nil {
		return err.Error
	}

	return nil
}
