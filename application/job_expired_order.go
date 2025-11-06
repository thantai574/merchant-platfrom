package application

import (
	"context"
	"go.uber.org/zap"
	"orders-system/domain/aggregates"
	"orders-system/domain/constants"
	"orders-system/proto/service_transaction"
)

func (us *OrderApplication) JobCancelExpiredOrder() {

	getExpiredOrder, err := us.OrderRepository.GetExpiredOrder(context.Background())
	if err != nil {
		us.Logger.Error("get_expired_order_err", zap.Error(err))
	} else {
		for index, v := range getExpiredOrder {
			if v.OrderType == constants.TRANSTYPE_WALLET_CASH_IN && v.BankCode == constants.GPNAPASGPBANK && !v.IsRetrievedOrder { //retrieveNAPAS
				getBankRetrieve, err := us.BankServiceRepository.RetrieveOrderStatus(v.OrderID, constants.GPNAPASGPBANK)
				if err != nil {
					v.IsRetrievedOrder = true
					us.Logger.With(zap.Error(err)).Error("bank_retrieve_er")

					failReason := getBankRetrieve.Message

					v.FailReason = failReason
					_, _ = us.FailedOrder(context.TODO(), v)

					getTranById, err := us.TransactionRepository.FindTransactionByID(context.TODO(), &service_transaction.ETransactionDTO{
						TransactionId: v.TransactionID,
					})
					if err == nil {
						getTranById.FailReason = failReason
						_, err = us.serviceTransactionCancel(context.TODO(), getTranById)
					}
				} else {
					v.IsRetrievedOrder = true
					getTranById, err := us.TransactionRepository.FindTransactionByID(context.TODO(), &service_transaction.ETransactionDTO{
						TransactionId: v.TransactionID,
					})
					if err == nil {
						if getBankRetrieve.ErrorCode.IsSuccess() {
							_, err = us.serviceTransactionConfirm(context.TODO(), getTranById)
							_, _ = us.SuccessOrder(context.TODO(), v)
						}
						if getBankRetrieve.ErrorCode.IsFail() {
							_, err = us.serviceTransactionCancel(context.TODO(), getTranById)
							_, _ = us.FailedOrder(context.TODO(), v)
						}
					}

				}
			} else {
				getExpiredOrder[index].IsExpired = true
				getExpiredOrder[index].FailReason = constants.MsgExpiredOrder + " - job"
				_, err := us.FailedOrder(context.Background(), getExpiredOrder[index])
				if err == nil {
					var eTransaction aggregates.TransactionDetail
					eTransaction.OrderId = v.OrderID
					eTransaction.ProviderMerchantId = v.SubscribeMerchantID
					eTransaction.Status = constants.TRANSACTION_STATUS_FAILED
					eTransaction.RefId = v.RefID
					eTransaction.SubTransactionType = v.SubOrderType
					eTransaction.SourceOfFund = v.SourceOfFund
					eTransaction.TransactionType = v.OrderType
					eTransaction.ServiceType = v.ServiceType

					if v.TransactionID != "" {
						getTranById, err := us.TransactionRepository.FindTransactionByID(context.TODO(), &service_transaction.ETransactionDTO{
							TransactionId: v.TransactionID,
						})

						if err == nil && getTranById.State == constants.TRANSACTION_STATE_INITIAL {
							getTranById.FailReason = constants.MsgExpiredOrder
							if getTranById.OrderId == "" {
								getTranById.OrderId = v.OrderID
							}

							_, _ = us.serviceTransactionCancel(context.TODO(), getTranById)
						}
					}

					_ = us.CreateMessageMqtt(context.TODO(), constants.TopicMQTTUpdateExpiredOrder, constants.MQTTEventBackground, constants.TopicMQTTUpdateExpiredOrder, eTransaction, false)
				}
			}

		}
	}
}
