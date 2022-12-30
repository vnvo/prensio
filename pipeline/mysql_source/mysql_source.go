package mysql_source

import (
	"context"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/siddontang/go-log/log"
	"github.com/vnvo/prensio/config"
	cdc "github.com/vnvo/prensio/pipeline/cdc_event"
)

type MySQLBinlogSource struct {
	fromGtid *mysql.GTIDSet
	config   *config.MySQLSourceConfig
	canal    *canal.Canal
	eventCh  chan<- cdc.CDCEvent
}

func NewMySQLBinlogSource(config *config.CDCConfig, eventCh chan<- cdc.CDCEvent) (MySQLBinlogSource, error) {
	mys := MySQLBinlogSource{
		nil,
		&config.Mysql,
		nil,
		eventCh,
	}

	err := mys.prepareCanal()
	if err != nil {
		return mys, err
	}

	return mys, nil
}

func (mys *MySQLBinlogSource) prepareCanal() error {
	conf := canal.NewDefaultConfig()
	conf.Addr = mys.config.Addr
	conf.User = mys.config.User
	conf.Password = mys.config.Pass
	conf.Charset = "utf8"
	conf.Flavor = "mysql"

	conf.ServerID = mys.config.ServerId
	conf.MaxReconnectAttempts = 5
	conf.Dump.ExecutionPath = ""
	conf.Dump.DiscardErr = false
	conf.Dump.SkipMasterData = true

	new_canal, err := canal.NewCanal(conf)
	if err != nil {
		panic(err)
	}

	mys.canal = new_canal
	mys.canal.SetEventHandler(&eventHandler{mys, nil})

	return nil
}

func (mys *MySQLBinlogSource) Init(lastState string) error {
	mys.fromGtid = nil
	if len(lastState) > 0 {
		gtidSet, err := mysql.ParseGTIDSet("mysql", lastState)
		if err != nil {
			log.Errorf("[%s] failed to parse last state: %v", mys.config.Name, err)
			return err
		}

		mys.fromGtid = &gtidSet
	}

	return nil
}

func (mys *MySQLBinlogSource) Run(ctx context.Context) error {
	if mys.fromGtid != nil {
		log.Info("starting binlog reader from GTID=%s", *mys.fromGtid)
		mys.canal.StartFromGTID(*mys.fromGtid)

	} else {
		log.Info("starting binlog reader from the begining (first GTID available)")
		mys.canal.Run()
	}

	<-ctx.Done()
	return nil
}

func (mys *MySQLBinlogSource) Close() {
	mys.canal.Close()
}

func (mys *MySQLBinlogSource) Query(query string) (*mysql.Result, error) {
	return mys.canal.Execute(query)
}
