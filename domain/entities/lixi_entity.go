package entities

import (
	"orders-system/proto/order_system"
	"time"
)

type LixiEntity struct {
	ID           string    `json:"id" bson:"_id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	MerchantID   string    `json:"merchant_id" bson:"merchant_id,omitempty"`
	Amount       int64     `json:"amount"`
	MaxUser      int64     `json:"max_user" bson:"max_user,omitempty"`
	Method       string    `json:"method"`
	AmountRemain int64     `json:"amount_remain" bson:"amount_remain,omitempty"`
	AmountUsed   int64     `json:"amount_used" bson:"amount_used,omitempty"`
	NumberScan   int64     `json:"number_scan" bson:"number_scan,omitempty"`
	StartDate    int64     `json:"start_date" bson:"start_date,omitempty"`
	FixAmount    int64     `json:"fix_amount" bson:"fixed_amount,omitempty"`
	MaxAmount    int64     `json:"max_amount" bson:"max_amount,omitempty"`
	MinAmount    int64     `json:"min_amount" bson:"min_amount,omitempty"`
	UserIds      []string  `json:"user_ids" bson:"user_ids,omitempty"`
	UserPhones   []string  `json:"phone_numbers" bson:"phone_numbers,omitempty"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at" bson:"updated_at,omitempty"`
}

func (lixi *LixiEntity) ConvertToProto() *order_system.Lixi {
	return &order_system.Lixi{
		ID:         lixi.ID,
		Name:       lixi.Name,
		Status:     lixi.Status,
		MerchantId: lixi.MerchantID,
		Amount:     lixi.Amount,
		MaxUser:    lixi.MaxUser,
		Method:     lixi.Method,
		StartDate:  lixi.StartDate,
		FixAmount:  lixi.FixAmount,
		MaxAmount:  lixi.MaxAmount,
		MinAmount:  lixi.MinAmount,
	}
}

func (lixi *LixiEntity) ConvertProtoToEntity(proto *order_system.Lixi) *LixiEntity {
	return &LixiEntity{
		ID:         proto.ID,
		Name:       proto.Name,
		Status:     proto.Status,
		MerchantID: proto.MerchantId,
		Amount:     proto.Amount,
		MaxUser:    proto.MaxUser,
		Method:     proto.Method,
		StartDate:  proto.StartDate,
		FixAmount:  proto.FixAmount,
		MaxAmount:  proto.MaxAmount,
		MinAmount:  proto.MinAmount,
		UserIds:    proto.UserIds,
	}
}
