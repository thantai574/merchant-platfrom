package aggregates

import "orders-system/domain/entities"

type OrderBuyCardDetail struct {
	Order *entities.OrderEntity
}
