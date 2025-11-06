package application

import (
	"context"
	"errors"
	"orders-system/domain/entities"
	"orders-system/proto/order_system"

	"go.uber.org/zap"
)

func (us *OrderApplication) GetDetailOrder(ctx context.Context, orderId string) (order_dto *entities.OrderEntity, err error) {
	order_dto = new(entities.OrderEntity)

	findOrderByIdResponse, err := us.OrderRepository.FindByOrderID(ctx, orderId)
	if err != nil || findOrderByIdResponse == nil {
		return order_dto, errors.New("Order is not exist")
	}

	order_dto = findOrderByIdResponse
	return order_dto, err
}

func (us *OrderApplication) GetDetailOrderByMerchantOrderId(ctx context.Context,
	req *order_system.GetDetailOrderRequest) (orderDto *entities.OrderEntity, err error) {
	orderDto = new(entities.OrderEntity)

	findOrderByIdResponse, err := us.OrderRepository.FindByMerchantIdAndRefId(ctx, req.MerchantId, req.RefId)
	if err != nil || findOrderByIdResponse == nil {
		us.Logger.Error("OrderRepository.FindByOrderID error", zap.Any("req", req))
		return orderDto, errors.New("Order is not exist")
	}

	orderDto = findOrderByIdResponse
	return orderDto, err
}
