package pipeline

import (
	"context"
	"sync"
	"time"

	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/siddontang/go-log/log"
	"github.com/vnvo/prensio/config"
	"github.com/vnvo/prensio/pipeline/cdc_event"
	"github.com/vnvo/prensio/pipeline/mysql_source"
	"github.com/vnvo/prensio/pipeline/transform"
)

type CDCPipeline struct {
	name       string
	config     *config.CDCConfig
	source     mysql_source.MySQLBinlogSource
	transf     *transform.Transform
	sink       *CDCKafkaSink
	rawEventCh chan cdc_event.CDCEvent
	wg         sync.WaitGroup
}

func NewCDCPipeline(name string, config *config.CDCConfig) CDCPipeline {

	rawEventCh := make(chan cdc_event.CDCEvent)

	mys, err := mysql_source.NewMySQLBinlogSource(config, rawEventCh)
	if err != nil {
		panic(err)
	}

	trn, err := transform.NewTransform(config)
	if err != nil {
		panic(err)
	}

	k := NewCDCKafkaSink(&config.KafkaSink)
	//create state manager

	return CDCPipeline{
		name,
		config,
		mys,
		trn,
		k,
		rawEventCh,
		sync.WaitGroup{},
	}
}

func (cdc *CDCPipeline) Init() error {
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

func (cdc *CDCPipeline) Close() {
	cdc.source.Close()
	cdc.sink.Close()
	cdc.wg.Done()
}

func (cdc *CDCPipeline) readFromHandler(ctx context.Context) {
	for {
		select {
		case e := <-cdc.rawEventCh:
			cdc.transf.Apply(&e)
			d, err := e.ToJson()

			cdc.sink.Write([]cdc_event.CDCEvent{e}, ctx)

			log.Debugf("after transform == json:%v - err:%v", d, err)
		case <-ctx.Done():
			return
		case <-time.After(time.Millisecond * 100):
		}
	}
}

func (cdc *CDCPipeline) Query(query string) (*mysql.Result, error) {
	return cdc.source.Query(query)
}
