package application

import (
	"context"
	"orders-system/proto/order_system"

	"go.uber.org/zap"
)

func (us *OrderApplication) OrderUpdate(ctx context.Context, request *order_system.UpdateOrderRequest,
	response *order_system.UpdateOrderResponse) (err error) {
	orderDto, err := us.GetValidOrder(ctx, request.OrderRequest.OrderId)
	if err != nil {
		us.Logger.Error("GetValidOrder error", zap.Any("request", request), zap.Error(err))
		return
	}

	// update order
	orderDto.ProtoToEntity(request.OrderRequest)
	orderDto, err = us.OrderRepository.ReplaceByID(ctx, orderDto)
	if err != nil {
		us.Logger.Error("OrderRepository.ReplaceByID error", zap.Any("request", request), zap.Error(err))
		return
	}

	response.OrderEntity = orderDto.ConvertToProto()
	return
}
