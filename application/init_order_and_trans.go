package application

import (
	"context"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
)

func (us *OrderApplication) InitOrderAndTrans(ctx context.Context, orderDto *entities.OrderEntity) (*entities.OrderEntity, error) {
	if orderDto.OrderType == constants.TRANSTYPE_PG_PAY_BY_PAY_INTERNATIONAL_CARD && orderDto.MerchantCategoryCode == "" {
		orderDto.MerchantCategoryType = "2"
	}

	initOrder, err := us.InitOrder(ctx, orderDto)
	return initOrder, err
}
