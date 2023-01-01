package pipeline

import (
	"fmt"
	"time"

	"github.com/siddontang/go-log/log"
	"github.com/vnvo/prensio/config"
)

type StateTracker struct {
	KafkaAddr string
	Topic     string
	State     *PipelineState
	Config    *config.StateTracking
}

type PipelineState struct {
	Gtid      string
	Name      string
	Timestamp time.Time
}

func NewStateTracker(pname string, conf *config.CDCConfig) (*StateTracker, error) {
	topic := fmt.Sprintf("prensio-%s-state", pname)
	if conf.State.Topic != "" && len(conf.State.Topic) > 0 {
		topic = conf.State.Topic
	}

	st := StateTracker{
		"",
		topic,
		&PipelineState{
			"fa632d10-864b-11ed-843e-0242c0a86002:27",
			pname,
			time.Now(),
		},
		&conf.State,
	}

	err := st.init()
	if err != nil {
		log.Errorf("[%s] state tracker init failed: %s", err)
		return nil, err
	}

	log.Infof("[%s] state tracker initialized. GTID=%s", pname, st.State.Gtid)
	return &st, nil
}

func (st *StateTracker) init() error {
	return nil
}

func (st *StateTracker) load() (string, error) {
	log.Debugf("[%s] loading last GTID", st.Topic)
	return "", nil
}

func (st *StateTracker) save(gtid string) error {
	log.Debugf("[%s] saving GTID=%s", st.Topic, gtid)
	return nil
}

func (st *StateTracker) Close() error {
	return nil
}
