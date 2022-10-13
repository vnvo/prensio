package pipeline

import (
	"context"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/vnvo/go-mysql-kafka/config"
)

type MySQLBinlogSource struct {
	config *config.MySQLSourceConfig
	canal  *canal.Canal
}

func NewMySQLBinlogSource(config *config.CDCConfig) (MySQLBinlogSource, error) {
	mys := MySQLBinlogSource{
		&config.Mysql,
		nil,
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
	return nil
}

func (mys *MySQLBinlogSource) Close() error {
	return nil
}
