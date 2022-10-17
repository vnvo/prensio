package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dbUser = "root"
	dbPass = "root"
	dbName = "main_test"
)

type mysqlContainer struct {
	container testcontainers.Container
	URI       string
}

func setupMySQL(ctx context.Context) (*mysqlContainer, error) {
	seedPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	contextPath := seedPath + "/.."
	fmt.Println(contextPath)

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    contextPath,
			Dockerfile: "Dockerfile.test",
		},
		WaitingFor: wait.ForLog("/usr/sbin/mysqld: ready for connections"),
	}

	mysqlC, err := testcontainers.GenericContainer(
		ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})

	if err != nil {
		panic(err)
	}

	host, _ := mysqlC.Host(ctx)
	p, _ := mysqlC.MappedPort(ctx, "6606/tcp")
	port := p.Int()

	var connUri = "%s:%s@tcp(%s:%d)/%s?tls=skip-verify&amp;parseTime=true&amp;multiStatement=true"
	connString := fmt.Sprintf(connUri, dbUser, dbPass, host, port, dbName)

	return &mysqlContainer{
		mysqlC,
		connString,
	}, nil

}

func TestMain(m *testing.M) {
	fmt.Println("starting the test")
	ctx := context.Background()

	//setup mysql
	mysqlC, err := setupMySQL(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(mysqlC)

	// make a connection
	// insert some data
	// validate the kafka message

	defer mysqlC.container.Terminate(ctx)

}
