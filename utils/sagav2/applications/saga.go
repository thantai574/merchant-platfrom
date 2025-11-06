package applications

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"orders-system/infrastructure/kafka"
	"orders-system/utils/configs"
	"orders-system/utils/gpooling"
	"orders-system/utils/helpers"
	"orders-system/utils/logger"
	"orders-system/utils/sagav2/domains"
	"time"
)

type Saga struct {
	ExecutionID   string
	steps         []*domains.Step
	Storage       kafka.Storage
	IPool         gpooling.IPool
	Logger        *zap.Logger
	Ctx           context.Context
	GroupConsumer sarama.ConsumerGroup
}

type Result struct {
	ExecutionError   error
	CompensateErrors []error
}

func NewCoordinator(id string, ctx context.Context, pool gpooling.IPool, conf *configs.Config, kafka_stg kafka.Storage) (*Saga, error) {
	log, _ := logger.NewLogger("production")
	if conf.KafkaConfig.Zookeepers == "" || conf.KafkaConfig.Brokers == "" {
		return nil, fmt.Errorf("Config don't have kafka ")
	}
	return &Saga{
		ExecutionID: id,
		Storage:     kafka_stg,
		IPool:       pool,
		Logger:      log,
		Ctx:         ctx,
	}, nil

}

func (sg *Saga) WithStep(step *domains.Step) *Saga {

	sg.steps = append(sg.steps, step)
	step.Index = len(sg.steps) - 1

	//Compensate Func Consumer
	sg.IPool.Submit(func() {
		cs, err := sg.Storage.ConsumePartition(sg.ExecutionID+"_"+fmt.Sprint(step.Index)+"_"+fmt.Sprint(domains.LogTypeSagaStepCompensate.ToString()), 0, sarama.OffsetOldest)
		if err != nil {
			sg.Logger.With(zapcore.Field{
				Key:       "ConsumePartition error",
				Type:      zapcore.ReflectType,
				Interface: err,
			})
			return
		}
		defer cs.Close()
		defer sg.Storage.Kazoo.DeleteTopic(sg.ExecutionID + "_" + fmt.Sprint(step.Index) + "_" + fmt.Sprint(domains.LogTypeSagaStepCompensate.ToString()))

		select {
		case msg := <-cs.Messages():
			var detail *domains.Log
			msgValue := string(msg.Value)
			sg.Logger.With(zapcore.Field{
				Key:       "msg",
				Type:      zapcore.ReflectType,
				Interface: msgValue,
			})
			err = json.Unmarshal([]byte(msgValue), &detail)

			err = step.CompensateFunc(sg.Ctx)
			if err == nil {
				sg.appendLog(&domains.Log{
					ExecutionID:  sg.ExecutionID,
					State:        domains.LogTypeSagaStepCompensate,
					Time:         helpers.GetCurrentTime(),
					StepNumber:   step.Index - 1,
					StepError:    detail.StepError,
					StepDuration: 0,
				})
			} else {
				sg.Logger.With(zapcore.Field{
					Key:       "err_compensate_saga",
					Type:      zapcore.ReflectType,
					Interface: err.Error(),
				})
			}
			if step.Index == 0 {
				sg.appendLog(&domains.Log{
					ExecutionID: sg.ExecutionID,
					State:       domains.LogTypeSagaComplete,
					Time:        helpers.GetCurrentTime(),
					StepNumber:  len(sg.steps),
					StepError:   detail.StepError,
				})
			}

		case <-time.After(time.Minute * 5):

		}

	})

	sg.IPool.Submit(func() {
		cs, err := sg.Storage.ConsumePartition(sg.ExecutionID+"_"+fmt.Sprint(step.Index)+"_"+fmt.Sprint(domains.LogTypeSagaStepExec.ToString()), 0, sarama.OffsetOldest)
		if err != nil {
			sg.Logger.With(zapcore.Field{
				Key:       "ConsumePartition error",
				Type:      zapcore.ReflectType,
				Interface: err,
			})
			return
		}
		defer cs.Close()
		defer sg.Storage.Kazoo.DeleteTopic(sg.ExecutionID + "_" + fmt.Sprint(step.Index) + "_" + fmt.Sprint(domains.LogTypeSagaStepExec.ToString()))

		select {
		case msg := <-cs.Messages():
			var detail *domains.Log
			msgValue := string(msg.Value)
			fmt.Println(msgValue)
			sg.Logger.With(zapcore.Field{
				Key:       "msg",
				Type:      zapcore.ReflectType,
				Interface: msgValue,
			})
			err = json.Unmarshal([]byte(msgValue), &detail)

			err = step.Func(sg.Ctx)
			lg := &domains.Log{
				ExecutionID: sg.ExecutionID,
				Time:        helpers.GetCurrentTime(),
			}
			if err != nil {
				sg.Logger.With(zapcore.Field{
					Key:       "err_step" + fmt.Sprint(step.Index) + "_saga" + sg.ExecutionID,
					Type:      zapcore.ReflectType,
					Interface: err.Error(),
				})
				lg.StepNumber = step.Index
				lg.StepError = err.Error()
				lg.State = domains.LogTypeSagaStepCompensate
				sg.appendLog(lg)
			} else {
				if step.Index == len(sg.steps)-1 {
					lg.State = domains.LogTypeSagaComplete
					lg.StepNumber = len(sg.steps)
				} else {
					lg.State = domains.LogTypeSagaStepExec
					lg.StepNumber = step.Index + 1
				}
				sg.appendLog(lg)
			}
		case <-time.After(time.Minute * 5):
		}

	})
	return sg
}

func (k *Saga) appendLog(log *domains.Log) error {
	log_json, err := json.Marshal(log)
	if err != nil {
		return err
	}
	_, _, err = k.Storage.SyncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: log.ExecutionID + "_" + fmt.Sprint(log.StepNumber) + "_" + fmt.Sprint(log.State.ToString()),
		Value: sarama.StringEncoder(string(log_json)),
	})

	return err
}

func (sg *Saga) Play() (res *Result) {
	var err_exec error
	sg.appendLog(&domains.Log{
		ExecutionID: sg.ExecutionID,
		State:       domains.LogTypeSagaStepExec,
		Time:        helpers.GetCurrentTime(),
		StepNumber:  0,
	})

	cs, err := sg.Storage.ConsumePartition(sg.ExecutionID+"_"+fmt.Sprint(len(sg.steps))+"_"+domains.LogTypeSagaComplete.ToString(), 0, sarama.OffsetOldest)

	if err != nil {
		return &Result{ExecutionError: err}
	}
	defer cs.Close()
	defer sg.Storage.Kazoo.DeleteTopic(sg.ExecutionID + "_" + fmt.Sprint(len(sg.steps)) + "_" + domains.LogTypeSagaComplete.ToString())

	select {
	case msg := <-cs.Messages():
		var detail *domains.Log
		msgValue := string(msg.Value)
		err = json.Unmarshal([]byte(msgValue), &detail)
		if detail.State == domains.LogTypeSagaComplete {
			if detail.StepError != "" {
				err_exec = fmt.Errorf(detail.StepError)
			}
		}

	case <-time.After(time.Minute * 5):
		err_exec = fmt.Errorf("Time out saga ")
	}

	return &Result{ExecutionError: err_exec}
}
