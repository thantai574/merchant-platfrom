package application

import (
	"context"
	"errors"
	"orders-system/domain/entities"
	internalErr "orders-system/errors"
)

func (us *OrderApplication) GetValidOrder(ctx context.Context, order_id string) (order_dto *entities.OrderEntity, err error) {
	getOrderById, err := us.OrderRepository.FindByOrderID(ctx, order_id)
	if err != nil || getOrderById == nil {
		return order_dto, errors.New("Order không hợp lệ")
	}
	if getOrderById.Status.IsVerifying() {
		return order_dto, internalErr.ErrPendingOrder
	}

	if getOrderById.Status.IsSuccess() {
		return order_dto, internalErr.ErrPaidOrder
	}

	if getOrderById.Status.IsFailed() {
		return order_dto, internalErr.ErrFailOrder
	}

	if getOrderById.IsExpired {
		return order_dto, internalErr.ErrExpiredOrder
	}

	return getOrderById, err
}
