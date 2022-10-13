package pipeline

import (
	"context"
	"fmt"

	"github.com/vnvo/go-mysql-kafka/config"
)

type CDCPipeline struct {
	name   string
	config *config.CDCConfig
	source MySQLBinlogSource

	//wg sync.WaitGroup
}

func NewCDCPipeline(name string, config *config.CDCConfig) CDCPipeline {
	//create mysql source
	mys, err := NewMySQLBinlogSource(config)
	if err != nil {
		panic(err)
	}

	//create transform
	//create kafka sink
	//create state manager

	return CDCPipeline{
		name,
		config,
		mys,
	}
}

func (cdc *CDCPipeline) Init() error {
	fmt.Println("init called on cdc-pipeline:", cdc.name)
	cdc.source.Init()

	return nil
}

func (cdc *CDCPipeline) Run(ctx context.Context) error {
	cdc.source.Run(ctx)
	return nil
}
