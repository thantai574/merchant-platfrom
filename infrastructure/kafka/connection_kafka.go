package kafka

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/lysu/kazoo-go"
	"strings"
	"time"
)

type Storage struct {
	sarama.Consumer
	sarama.SyncProducer
	sarama.AsyncProducer
	*kazoo.Kazoo
}

func NewConnection(ctx context.Context, zkAddrs, brokers string) (storage Storage, err error) {

	conf := kazoo.NewConfig()
	conf.Timeout = time.Minute

	kz, err := kazoo.NewKazoo(strings.Split(zkAddrs, ","), conf)

	if err != nil {
		panic(err)
	}

	consumer, err := sarama.NewConsumer(strings.Split(brokers, ","), nil)

	if err != nil {
		panic(err)
	}

	producer, err := sarama.NewSyncProducer(strings.Split(brokers, ","), nil)

	if err != nil {
		panic(err)
	}

	asyncProducer, err := sarama.NewAsyncProducer(strings.Split(brokers, ","), nil)

	if err != nil {
		panic(err)
	}

	return Storage{
		Kazoo:         kz,
		Consumer:      consumer,
		SyncProducer:  producer,
		AsyncProducer: asyncProducer,
	}, err

}
