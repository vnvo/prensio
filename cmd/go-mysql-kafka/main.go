package main

import (
	"context"
	"fmt"

	"github.com/vnvo/go-mysql-kafka/config"
	"github.com/vnvo/go-mysql-kafka/pipeline"
)

func main() {
	conf := config.NewCDCConfig("./go_mysql_kafka.toml")
	fmt.Println(conf)

	myPipeline := pipeline.NewCDCPipeline("first-pipeline", &conf)
	fmt.Println(myPipeline)

	err := myPipeline.Init()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	myPipeline.Run(ctx)
}
