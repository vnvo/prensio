package main

import (
	"context"
	"os"

	"github.com/siddontang/go-log/log"
	"github.com/vnvo/go-mysql-kafka/config"
	"github.com/vnvo/go-mysql-kafka/pipeline"
)

func main() {
	initLogger()

	conf := config.NewCDCConfig("./go_mysql_kafka.toml")

	myPipeline := pipeline.NewCDCPipeline("first-pipeline", &conf)
	log.Infof("[%s] created.", "first-pipeline")

	err := myPipeline.Init()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	myPipeline.Run(ctx)
}

func initLogger() {
	logLevel, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		logLevel = "debug"
	}

	log.SetLevelByName(logLevel)
}
