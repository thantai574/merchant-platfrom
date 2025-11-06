package mqtt

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
	"sync"
	"time"
)

func Connection(uri, user, password string) mqtt.Client {

	opts := mqtt.NewClientOptions().AddBroker(uri)
	opts.SetUsername(user)
	opts.SetPassword(password)
	opts.SetClientID(fmt.Sprint(time.Now().Add(time.Hour * 7).Unix()))

	client_mqtt := mqtt.NewClient(opts)

	if token := client_mqtt.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	return client_mqtt
}

type repositoryImpl struct {
	client []mqtt.Client
	zap.Logger
}

func NewMQTTRepositoryImpl(client []mqtt.Client, logger *zap.Logger) *repositoryImpl {
	return &repositoryImpl{client, *logger}
}

func (r repositoryImpl) Publish(topic, message string, retain bool, prefix string) (err error) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() { // mobile
		publish := r.client[0].Publish("3ojTg-NvIc0eQ7wGuXnTIlB-eiw3TN8J"+"/topic/"+topic+"/", byte(2), retain, message)
		if publish.Error() != nil {
			r.Logger.With(zap.Any("message", message)).
				With(zap.Any("topic", topic)).
				Error("MQTT_PUBLISH")
		}
		wg.Done()
	}()

	go func() { // internal client
		publish := r.client[1].Publish(prefix+"/topic/"+topic+"/", byte(2), retain, message)
		if publish.Error() != nil {
			r.Logger.With(zap.Any("message", message)).
				With(zap.Any("topic", topic)).
				Error("MQTT_PUBLISH")
		}
		wg.Done()
	}()

	wg.Wait()

	return err
}

func (r repositoryImpl) Subscribe(topic string, c func(client mqtt.Client, message mqtt.Message)) {
	for _, v := range r.client {
		v.Subscribe("3ojTg-NvIc0eQ7wGuXnTIlB-eiw3TN8J/topic/#", 0, c)
	}
}
