package pipeline

import (
	"context"
	"fmt"
	"sync"

	"github.com/vnvo/go-mysql-kafka/config"
)

type CDCPipeline struct {
	name   string
	config *config.CDCConfig
	source MySQLBinlogSource

	wg sync.WaitGroup
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
		sync.WaitGroup{},
	}
}

func (cdc *CDCPipeline) Init() error {
	fmt.Println("init called on cdc-pipeline:", cdc.name)
	cdc.source.Init()

	return nil
}

func (cdc *CDCPipeline) Run(ctx context.Context) error {
	cdc.wg.Add(1)
	go func() {
		defer cdc.wg.Done()
		cdc.source.Run(ctx, "")
	}()

	select {
	case <-ctx.Done():
		fmt.Println("pipeline.Run -> Done.")
	}

	return nil
}
