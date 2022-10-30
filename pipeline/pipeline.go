package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vnvo/go-mysql-kafka/cdc_event"
	"github.com/vnvo/go-mysql-kafka/config"
	"github.com/vnvo/go-mysql-kafka/mysql_source"
	"github.com/vnvo/go-mysql-kafka/transform"
)

type CDCPipeline struct {
	name   string
	config *config.CDCConfig
	source mysql_source.MySQLBinlogSource
	transf *transform.Transform

	rawEventCh chan cdc_event.CDCEvent
	wg         sync.WaitGroup
}

func NewCDCPipeline(name string, config *config.CDCConfig) CDCPipeline {

	rawEventCh := make(chan cdc_event.CDCEvent)

	//create mysql source
	mys, err := mysql_source.NewMySQLBinlogSource(config, rawEventCh)
	if err != nil {
		panic(err)
	}

	trn, err := transform.NewTransform(config)
	if err != nil {
		panic(err)
	}

	//create kafka sink
	//create state manager

	return CDCPipeline{
		name,
		config,
		mys,
		trn,
		rawEventCh,
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
		cdc.source.Run(ctx)
	}()

	cdc.wg.Add(1)
	go func() {
		defer cdc.wg.Done()
		cdc.readFromHandler(ctx)
	}()

	cdc.wg.Wait()

	return nil
}

func (cdc *CDCPipeline) readFromHandler(ctx context.Context) {
	for {
		select {
		case e := <-cdc.rawEventCh:
			fmt.Println(e.ToJson())
			cdc.transf.Apply(&e)
			fmt.Println(" == after transform ==")
			fmt.Println(e.ToJson())
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 1):
			//fmt.Println("just waiting for other channels ...")
		}
	}
}
