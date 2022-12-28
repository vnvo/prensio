package pipeline

import (
	"context"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/siddontang/go-log/log"
	"github.com/vnvo/prensio/config"
	"github.com/vnvo/prensio/pipeline/cdc_event"
)

type CDCKafkaSink struct {
	conf *config.KafkaSink
	w    *kafka.Writer
}

func NewCDCKafkaSink(conf *config.KafkaSink) *CDCKafkaSink {
	addrs := conf.GetAddrList() // []string{"localhost:29092", "localhost:39092", "localhost:49092"}

	k := kafka.Writer{
		Addr:                   kafka.TCP(addrs...),
		RequiredAcks:           1,
		AllowAutoTopicCreation: true,
		BatchTimeout:           time.Millisecond * 5,
		WriteBackoffMin:        time.Millisecond * 5,
		WriteBackoffMax:        time.Millisecond * 100,
		BatchBytes:             10240,
	}

	return &CDCKafkaSink{
		conf,
		&k,
	}
}

func (k *CDCKafkaSink) Close() {
	k.w.Close()
}

func (k *CDCKafkaSink) Write(msgs []cdc_event.CDCEvent, ctx context.Context) error {
	kmsgs := []kafka.Message{}

	for _, msg := range msgs {
		payload, _ := msg.ToJson()
		kmsgs = append(kmsgs, kafka.Message{
			Topic: msg.Kafka.Topic,
			Key:   []byte(msg.Kafka.Key),
			Value: []byte(payload),
		})
	}

	var err error
	for retry := 1; retry <= 5; retry += 1 {
		err = k.w.WriteMessages(ctx, kmsgs...)
		if err != nil {
			log.Errorf("write to kafka faild(try %d of 5). %v", retry, err)
			if strings.HasPrefix(err.Error(), "[5] Leader Not Available") {
				time.Sleep(time.Second * 3)
			} else {
				time.Sleep(time.Millisecond * 500)
			}
			log.Info("kafka next try ...")
			continue
		} else {
			log.Info("kafka write successful. duration=", time.Now().UnixMicro()-msgs[0].Meta.Timestamp)
			return nil
		}
	}

	return err
}
