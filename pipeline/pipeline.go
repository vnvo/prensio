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
	state      *StateTracker
	source     mysql_source.MySQLBinlogSource
	transf     *transform.Transform
	sink       *CDCKafkaSink
	rawEventCh chan cdc_event.CDCEvent
	wg         sync.WaitGroup
}

func NewCDCPipeline(name string, conf *config.CDCConfig) CDCPipeline {

	st, err := NewStateTracker(name, conf)
	if err != nil {
		log.Errorf("[%s] unable to track the state ... stopping.", name)
		panic("state tracking failed")
	}

	rawEventCh := make(chan cdc_event.CDCEvent)

	mys, err := mysql_source.NewMySQLBinlogSource(conf, rawEventCh)
	if err != nil {
		panic(err)
	}

	trn, err := transform.NewTransform(conf)
	if err != nil {
		panic(err)
	}

	k := NewCDCKafkaSink(&conf.KafkaSink)
	//create state manager

	return CDCPipeline{
		name,
		conf,
		st,
		mys,
		trn,
		k,
		rawEventCh,
		sync.WaitGroup{},
	}
}

func (cdc *CDCPipeline) Init() error {
	lastState := cdc.state.State.Gtid
	err := cdc.source.Init(lastState)
	if err != nil {
		panic(err)
	}

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
	log.Info("closing the pipeline")
	cdc.source.Close()
	cdc.sink.Close()
	cdc.wg.Done()
}

func (cdc *CDCPipeline) readFromHandler(ctx context.Context) {
	for {
		select {
		case e := <-cdc.rawEventCh:
			tStart := time.Now().UnixMicro()
			verdict, err := cdc.transf.Apply(&e)
			log.Debugf("transform duration(us)=%d", time.Now().UnixMicro()-tStart)
			if err != nil {
				log.Errorf("transform failed: %v", err)
				cdc.Close()
				return
			}

			if verdict == transform.ACTION_CONT {
				wStart := time.Now().UnixMicro()
				d, err := e.ToJson()
				if err != nil {
					log.Errorf(
						"unable to convert to json ... stoppping. %s",
						e.String(),
					)
					panic("unable to encode the event")
				}

				err = cdc.sink.Write([]cdc_event.CDCEvent{e}, ctx)
				if err != nil {
					panic("unable to write to kafka")
				}
				cdc.state.save(e.GetGTID())
				log.Debugf("read to sink duration(us)=%d", time.Now().UnixMicro()-wStart)
				log.Debugf("===\n%s\n\nJSON: %v\n\nError: %v", e.String(), d, err)
			}

		case <-ctx.Done():
			return
		case <-time.After(time.Millisecond * 100):
		}
	}
}

func (cdc *CDCPipeline) Query(query string) (*mysql.Result, error) {
	return cdc.source.Query(query)
}
