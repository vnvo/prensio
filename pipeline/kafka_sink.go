package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/siddontang/go-log/log"
	"github.com/vnvo/go-mysql-kafka/config"
	"github.com/vnvo/go-mysql-kafka/pipeline/cdc_event"
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
	}

	return &CDCKafkaSink{
		conf,
		&k,
	}
}

func (k *CDCKafkaSink) Write(msgs []cdc_event.CDCEvent, ctx context.Context) error {
	kmsgs := []kafka.Message{}

	for _, msg := range msgs {
		payload, _ := msg.ToJson()
		kmsgs = append(kmsgs, kafka.Message{
			Topic: fmt.Sprintf("%s.%s", msg.Schema, msg.Table),
			Key:   []byte("cdc_event"),
			Value: []byte(payload),
		})
	}

	for retry := 1; retry <= 3; retry += 1 {
		err := k.w.WriteMessages(ctx, kmsgs...)
		if err != nil {
			log.Errorf("write to kafka faild(try %d of 3). %v", retry, err)
			time.Sleep(time.Second)
			continue
		} else {
			break
		}
	}

	return nil
}
