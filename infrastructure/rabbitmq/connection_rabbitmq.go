package rabbitmq

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"orders-system/domain/entities"
	"orders-system/utils/configs"
	"orders-system/utils/gpooling"
)

type options struct {
	Uri        string
	AutoAck    bool
	AutoDelete bool
	Durable    bool
	Exclusive  bool
	NoWait     bool
}

func NewOptions() *options {
	return &options{}
}

func (o *options) WithUri(uri string) *options {
	o.Uri = uri
	return o
}

func (o *options) WithAutoAck(ack bool) *options {
	o.AutoAck = ack
	return o
}

type RabbiMQ struct {
	Connection *amqp.Connection
	IPool      gpooling.IPool
	options
	configs.Config
	*zap.Logger
}

func NewRabbiMQ(o options, conf configs.Config, log *zap.Logger, pool gpooling.IPool) (*RabbiMQ, error) {
	conn, err := amqp.Dial(o.Uri)

	if err != nil {
		panic(err)
	}

	return &RabbiMQ{
		IPool:      pool,
		Connection: conn,
		options:    o,
		Config:     conf,
		Logger:     log,
	}, nil
}
func (r *RabbiMQ) PublishToExchange(msg proto.Message, topic string) error {
	ch, err := r.Connection.Channel()

	if err != nil {
		return err
	}

	send_data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	err = ch.Publish(
		topic, // exchange
		topic, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        send_data,
		})

	return err

}

func (r *RabbiMQ) WithConsumerQueue(fn func(msg []byte) error, queue_name string, retry bool) error {
	r.IPool.Submit(func() {
		ch, err := r.Connection.Channel()
		defer ch.Close()
		if err != nil {
			r.Logger.With(zap.Field{
				Key:       "err-msg-queue-" + queue_name,
				Type:      zapcore.ReflectType,
				Interface: err,
			}).Info("err queue ")
			return
		}
		q, err := ch.QueueDeclare(
			queue_name,   // name
			true,         // durable
			r.AutoDelete, // delete when usused
			r.Exclusive,  // exclusive
			r.NoWait,     // no-wait
			nil,          // arguments
		)
		if err != nil {
			r.Logger.With(zap.Field{
				Key:       "err-msg-queue-" + queue_name,
				Type:      zapcore.ReflectType,
				Interface: err,
			}).Info("err queue ")
			return
		}

		msgs, err := ch.Consume(
			q.Name,      // queue
			"",          // consumer
			r.AutoAck,   // auto-ack
			r.Exclusive, // exclusive
			false,       // no-local
			false,       // no-wait
			nil,         // args
		)

		if err != nil {
			r.Logger.With(zap.Field{
				Key:       "err-msg-queue-" + queue_name,
				Type:      zapcore.ReflectType,
				Interface: err,
			}).Info("err queue ")
			return
		}

		for d := range msgs {
			fn(d.Body)
			if retry {
				_ = d.Ack(false)
			} else {
				_ = d.Ack(true)
			}

		}

		return
	})

	return nil
}

func (r *RabbiMQ) WithConsumerTopic(fn func(ctx context.Context, msg []byte) (eOrder *entities.OrderEntity, err error), topic_name string) (err error) {
	r.IPool.Submit(func() {
		ch, err := r.Connection.Channel()
		defer ch.Close()
		if err != nil {
			r.Logger.With(zap.Field{
				Key:       "err-msg-topic-queue-" + topic_name,
				Type:      zapcore.ReflectType,
				Interface: err,
			}).Info("err queue ")
			return
		}

		err = ch.ExchangeDeclare(
			topic_name, // name
			"topic",
			r.Durable,   // durable
			true,        // delete when usused
			r.Exclusive, // exclusive
			r.NoWait,    // no-wait
			nil,         // arguments
		)

		q, err := ch.QueueDeclare(
			topic_name,   // name
			true,         // durable
			r.AutoDelete, // delete when usused
			r.Exclusive,  // exclusive
			r.NoWait,     // no-wait
			nil,          // arguments
		)
		if err != nil {
			r.Logger.With(zap.Field{
				Key:       "err-msg-queue-" + topic_name,
				Type:      zapcore.ReflectType,
				Interface: err,
			}).Info("err queue ")
			return
		}

		err = ch.QueueBind(
			q.Name,          // queue name
			topic_name+".#", // routing key
			topic_name,      // exchange
			false,
			nil,
		)

		if err != nil {
			r.Logger.With(zap.Field{
				Key:       "err-msg-queue-" + topic_name,
				Type:      zapcore.ReflectType,
				Interface: err,
			}).Info("err queue ")
			return
		}

		msgs, err := ch.Consume(
			q.Name,      // queue
			"",          // consumer
			false,       // auto-ack
			r.Exclusive, // exclusive
			false,       // no-local
			false,       // no-wait
			nil,         // args
		)

		if err != nil {
			r.Logger.With(zap.Field{
				Key:       "err-msg-queue-" + topic_name,
				Type:      zapcore.ReflectType,
				Interface: err,
			}).Info("err queue ")
			return
		}

		for d := range msgs {
			_, err = fn(context.TODO(), d.Body)
			ch.Ack(d.DeliveryTag, false)
		}

		return
	})

	return err
}
