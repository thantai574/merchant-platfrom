package application

import (
	"context"
	"orders-system/domain/aggregates"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	"orders-system/proto/order_system"
	"orders-system/proto/service_transaction"
)

func (us *OrderApplication) OrderCancel(ctx context.Context, request *order_system.CancelOrderRequest, response *order_system.CancelOrderResponse) (orderDto *entities.OrderEntity, err error) {
	getOrderById, err := us.GetValidOrder(ctx, request.OrderId)
	if err != nil {
		return orderDto, err
	}

	orderDto, err = us.CancelOrder(ctx, getOrderById)
	if err != nil {
		return orderDto, err
	}

	if orderDto.TransactionID != "" {
		findTrans, err := us.TransactionRepository.FindTransactionByID(ctx, &service_transaction.ETransactionDTO{
			TransactionId: orderDto.TransactionID,
		})

		if err == nil && findTrans.State == constants.TRANSACTION_STATE_INITIAL {
			findTrans.OrderId = orderDto.OrderID
			_, _ = us.serviceTransactionCancel(ctx, findTrans)
		}
	}

	response.OrderEntity = orderDto.ConvertToProto()

	eTransaction := aggregates.TransactionDetail{
		Transaction: entities.Transaction{
			OrderId:            orderDto.OrderID,
			TransactionType:    orderDto.OrderType,
			SubTransactionType: orderDto.SubOrderType,
			SourceOfFund:       orderDto.SourceOfFund,
			ServiceType:        orderDto.ServiceType,
		},
	}

	us.IPool.Submit(func() {
		_ = us.CreateMessageMqtt(ctx, constants.TopicUpdateCancelOrder, constants.MQTTEventBackground, constants.TopicUpdateCancelOrder, eTransaction, false)
	})

	return orderDto, err
}
