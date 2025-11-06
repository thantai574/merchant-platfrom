package application

import (
	"context"
	"encoding/json"
	"orders-system/domain/constants"
	"orders-system/proto/service_transaction"
)

func (us *OrderApplication) CreateMessageMqtt(ctx context.Context, topic, event, key string,
	data interface{}, retain bool) error {
	message := new(constants.Message)

	message.Event = event
	message.Key = key
	message.MessageData = data

	prefix := us.Config.MQTTInternalUri.Prefix
	json_send, err := json.Marshal(message)

	if err != nil {
		//us.Logger.Warn("[CreateMessageMqtt] - Can not marshal request")
		return err
	}
	err = us.MQTT.Publish(topic, string(json_send), retain, prefix)
	if err != nil {
		//us.Logger.Warn("[CreateMessageMqtt] - Can not publish msg mqtt, topic ", zap.String("topic", topic), zap.String("data", string(json_send)))
		return err
	}
	//us.Logger.Warn("[CreateMessageMqtt] - Succss publish msg mqtt, topic: %s, data: %s", zap.String("topic", topic), zap.String("data", string(json_send)))
	return nil
}

func (us *OrderApplication) MqttUpdateProfile(ctx context.Context, event, payerId, payeeId string) (err error) {

	if payerId != "" {
		payer, err := us.GetProfile(ctx, payerId)
		if err == nil {
			us.CreateMessageMqtt(context.TODO(), payer.Id, constants.MQTTEventBackground, event, payer, false)
		}
	}

	if payeeId != "" {
		payee, err := us.GetProfile(ctx, payeeId)

		if err == nil {
			_ = us.CreateMessageMqtt(context.TODO(), payee.Id, constants.MQTTEventBackground, event, payee, false)
		}
	}

	return
}

func (us *OrderApplication) SendMqttTransactionByObject(ctx context.Context, event string, trans service_transaction.ETransactionDTO) (err error) {
	transaction, _ := us.ConvertETransToDetail(ctx, trans)
	if trans.PayerId != "" {
		us.CreateMessageMqtt(context.TODO(), transaction.PayerId, constants.MQTTEventBackground, event, transaction, false)
	}

	if trans.PayeeId != "" {
		us.CreateMessageMqtt(context.TODO(), transaction.PayeeId, constants.MQTTEventBackground, event, transaction, false)
	}

	if transaction.Status == constants.TRANSACTION_STATUS_FINISH && transaction.Type != constants.TransactionTypeCashback {
		us.CreateMessageMqtt(context.TODO(), transaction.PayeeId, constants.MQTTEventNotification, constants.MQTTTransactionSuccess, transaction, false)

		us.CreateMessageMqtt(context.TODO(), transaction.PayerId, constants.MQTTEventNotification, constants.MQTTTransactionSuccess, transaction, false)
		if transaction.TransactionType == constants.TRANSTYPE_WALLET_REFUND {
			us.CreateMessageMqtt(context.TODO(), transaction.UserReceiveRefund, constants.MQTTEventNotification, constants.MQTTTransactionSuccess, transaction, false)
		}
	}

	if transaction.Status == constants.TRANSACTION_STATUS_FAILED {
		if trans.PayerId != "" {
			us.CreateMessageMqtt(context.TODO(), transaction.PayerId, constants.MQTTEventNotification, constants.MQTTTransactionFailed, transaction, false)
		}

		if trans.PayeeId != "" {
			us.CreateMessageMqtt(context.TODO(), transaction.PayeeId, constants.MQTTEventNotification, constants.MQTTTransactionFailed, transaction, false)
		}
	}

	if transaction.Type == constants.TRANSTYPE_WALLET_LIXI && transaction.Status == constants.TRANSACTION_STATUS_FINISH {
		us.CreateMessageMqtt(context.TODO(), transaction.PayeeId, constants.MQTTEventNotification, constants.MQTTLIXISuccess, transaction, false)

	}
	return
}
