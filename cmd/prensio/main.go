package main

import (
	"context"
	"os"

	"github.com/siddontang/go-log/log"
	"github.com/vnvo/prensio/config"
	"github.com/vnvo/prensio/pipeline"
)

func main() {
	initLogger()

	conf, err := config.NewCDCConfig("./go_mysql_kafka.toml")
	if err != nil {
		panic(err)
	}

	myPipeline := pipeline.NewCDCPipeline("first-pipeline", &conf)
	log.Infof("[%s] created.", "first-pipeline")

	err = myPipeline.Init()
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
