package entities

import "orders-system/proto/order_system"

type EntityStatus order_system.OrderStatus

func (o EntityStatus) StatusOrderProto() order_system.OrderStatus {
	return order_system.OrderStatus(o)
}

func (o *EntityStatus) ConvertStatusOrderEntity(i order_system.OrderStatus) EntityStatus {
	*o = EntityStatus(i)
	return *o
}
func (o EntityStatus) StatusString() (out string) {
	if data, ok := order_system.OrderStatus_name[int32(o)]; ok == true {
		return data
	}
	return
}
func (o EntityStatus) IsProcessing() bool {
	return o.StatusOrderProto() == order_system.OrderStatus_ORDER_PROCESSING
}

func (o EntityStatus) IsPending() bool {
	return o.StatusOrderProto() == order_system.OrderStatus_ORDER_PENDING
}

func (o EntityStatus) IsSuccess() bool {
	return o.StatusOrderProto() == order_system.OrderStatus_ORDER_SUCCESS
}

func (o EntityStatus) IsFailed() bool {
	return o.StatusOrderProto() == order_system.OrderStatus_ORDER_FAILED
}
func (o EntityStatus) IsVerifying() bool {
	return o.StatusOrderProto() == order_system.OrderStatus_ORDER_VERIFYING
}
func (o EntityStatus) IsCancel() bool {
	return o.StatusOrderProto() == order_system.OrderStatus_ORDER_CANCEL
}
