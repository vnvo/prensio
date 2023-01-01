package mysql_source

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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
		//lastState = mys.getGTIDRange(masterGTID, lastGTID string)
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
	masterGTID, err := mys.canal.GetMasterGTIDSet()
	if err != nil {
		panic(err)
	}

	log.Debugf("master GTID=%s", masterGTID)

	if mys.fromGtid != nil {
		lastGTID := *mys.fromGtid
		gtidStart := mys.getGTIDRange(masterGTID.String(), lastGTID.String())
		startSet, err := mysql.ParseGTIDSet("mysql", gtidStart)
		if err != nil {
			log.Errorf("[%s] failed to parse calculated start state: %v", mys.config.Name, err)
			return err
		}

		log.Infof("starting binlog reader from GTID=%s, calculated start=%s", lastGTID, gtidStart)
		mys.canal.StartFromGTID(startSet)

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

func (mys *MySQLBinlogSource) getGTIDRange(masterGTID, lastGTID string) string {
	// master=fa632d10-864b-11ed-843e-0242c0a86002:1-28
	// lastgtid=fa632d10-864b-11ed-843e-0242c0a86002:18
	log.Debugf("masterGTID=%s lastGTID=%s", masterGTID, lastGTID)
	r := regexp.MustCompile("(.*):(.*)-(.*)")
	master := r.FindStringSubmatch(masterGTID)[1:] //uuid, start, end
	last := strings.Split(lastGTID, ":")           //uuid, end

	if master[0] == last[0] {
		if master[2] >= last[1] {
			//master is ahead, safe to use the last
			return fmt.Sprintf("%s:%s-%s", master[0], master[1], last[1])
		} else {
			// master is behind, this probably needs manual interventaion or
			// some kind of pre-defined recovery strategy.
			// for now, lets panic.
			log.Error(
				"[%s] last GTID is ahead of current master GTID\nMaster GTID=%s, Prensio GTID=%s",
				mys.config.Name, masterGTID, lastGTID,
			)
			log.Errorf("Tip: Consider manually adjusting the pipeline state. Side effects: duplicate events!")
			panic("")
		}

	} else {
		log.Errorf(
			"[%s] GTID mismatch. master=%s, prensio=%s",
			mys.config.Name, masterGTID, lastGTID)
		return lastGTID
	}

}
