package mysql_source

import (
	"context"

	"github.com/go-mysql-org/go-mysql/canal"
	cdc "github.com/vnvo/go-mysql-kafka/cdc_event"
	"github.com/vnvo/go-mysql-kafka/config"
)

type MySQLBinlogSource struct {
	config  *config.MySQLSourceConfig
	canal   *canal.Canal
	eventCh chan<- cdc.CDCEvent
}

func NewMySQLBinlogSource(config *config.CDCConfig, eventCh chan<- cdc.CDCEvent) (MySQLBinlogSource, error) {
	mys := MySQLBinlogSource{
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
	mys.canal.SetEventHandler(&eventHandler{mys})

	return nil
}

func (mys *MySQLBinlogSource) Init() error {
	return nil
}

func (mys *MySQLBinlogSource) Run(ctx context.Context) error {
	mys.canal.Run()
	<-ctx.Done()

	return nil
}

func (mys *MySQLBinlogSource) Close() error {
	return nil
}
