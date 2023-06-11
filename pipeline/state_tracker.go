package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/siddontang/go-log/log"
	"github.com/vnvo/prensio/config"
)

type StateTracker struct {
	State  *PipelineState
	Config *config.StateTracking
	conn   *kafka.Conn
}

type PipelineState struct {
	Gtid      string `json:"gtid"`
	Name      string `json:"name"`
	Timestamp int64  `json:"timestamp"`
}

func NewStateTracker(pname string, conf *config.CDCConfig) (*StateTracker, error) {
	topic := fmt.Sprintf("prensio-%s-state", pname)
	if strings.TrimSpace(conf.State.Topic) == "" {
		conf.State.Topic = topic
	}

	log.Infof("state tracker init ... topic=%s | conf.topic=%s", topic, conf.State.Topic)

	st := StateTracker{
		&PipelineState{
			"", //"fa632d10-864b-11ed-843e-0242c0a86002:27",
			pname,
			time.Now().UnixMicro(),
		},
		&conf.State,
		nil,
	}

	err := st.init()
	if err != nil {
		log.Errorf("[%s] state tracker init failed: %s", pname, err)
		return nil, err
	}

	log.Infof("[%s] state tracker initialized. GTID=%s", pname, st.State.Gtid)
	return &st, nil
}

func (st *StateTracker) init() error {
	/*
		- setup kafka client
		- read the last state from state topic
		- pass the last state along
	*/
	log.Infof(
		"[%s] state tracker init ... topic=%s",
		st.State.Name, st.Config.Topic)

	var err error
	st.conn, err = kafka.DialLeader(
		context.Background(), "tcp", st.Config.KafkaAddr,
		st.Config.Topic, 0)
	if err != nil {
		log.Fatal("failed to dial kafka:", err)

	}

	st.ensureTopicExists()
	st.load()

	return nil
}

func (st *StateTracker) load() error {
	log.Debugf("[%s] loading last GTID", st.Config.Topic)
	// get the most recent message
	offset, whence := st.conn.Offset()
	lastOffset, err := st.conn.ReadLastOffset()

	log.Infof("state offset=%d, whence=%d", offset, whence)
	log.Infof("state lastOffset=%d, err=%d", lastOffset, err)

	if lastOffset > 0 { // there is something to load
		if offset, err := st.conn.Seek(1, kafka.SeekEnd); err != nil {
			log.Fatal("unable to read the latest state", err)
		} else {
			log.Infof("loading state at offset=%d", offset)

			msg, err := st.conn.ReadMessage(10240)

			if err != nil {
				log.Fatal("unable to read the latest state", err)
			}

			log.Infof("last state = %v", msg)
			err = json.Unmarshal(msg.Value, st.State)
			if err != nil {
				log.Fatalf(
					"[%s] failed to restore state from latest kafka message: %v",
					st.State.Name, nil)
			}

		}
	}

	return nil
}

func (st *StateTracker) save(gtid string) error {
	if len(gtid) > 0 {
		st.State.Gtid = gtid
		st.State.Timestamp = time.Now().UnixMicro()
	}

	log.Debugf("[%s] saving GTID=%s", st.Config.Topic, st.State.Gtid)
	stateJson, err := json.Marshal(st.State)

	if err != nil {
		return err
	}

	kmsgs := []kafka.Message{
		{
			Topic: st.Config.Topic,
			Key:   []byte(st.State.Name),
			Value: []byte(stateJson),
		},
	}

	for retry := 1; retry <= 5; retry += 1 {
		_, err = st.conn.WriteMessages(kmsgs...)
		if err != nil {
			log.Errorf("[%s] writing state to kafka faild(try %d of 5). %v", st.State.Name, retry, err)
			if strings.HasPrefix(err.Error(), "[5] Leader Not Available") {
				time.Sleep(time.Second * 3)
			} else {
				time.Sleep(time.Millisecond * 500)
			}
			log.Info("kafka next try ...")
			continue
		}

		break
	}

	return nil
}

func (st *StateTracker) close() {
	st.save("")
	st.conn.Close()
}

func (st *StateTracker) ensureTopicExists() error {
	tp := st.getTopic()
	if tp == nil {
		tp = st.createTopic()
	}

	log.Infof(
		"state topic is available. topic=%s, Isr=%d, leader-id=%d",
		st.Config.Topic, len(tp.Isr), tp.Leader.ID)

	return nil
}

func (st *StateTracker) getTopic() *kafka.Partition {
	partitions, err := st.conn.ReadPartitions(st.Config.Topic)
	if err != nil {
		log.Fatal(err)
	}

	if len(partitions) != 1 {
		log.Fatalf(
			"state topic should only have one partition. topic=%s partitions=%d",
			st.Config.Topic, len(partitions))
	}

	return &partitions[0]
}

func (st *StateTracker) createTopic() *kafka.Partition {
	return nil
}
