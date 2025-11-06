package entities

import (
	"github.com/golang/protobuf/ptypes"
	"orders-system/proto/order_system"
	"time"
)

type VirtualAccounts struct {
	Id                 string    `json:"id" bson:"_id,omitempty"`
	Status             string    `json:"status" bson:"status,omitempty"`
	Provider           string    `json:"provider" bson:"provider,omitempty"`
	Balance            int64     `json:"balance" bson:"balance,omitempty"`
	AccountNunmber     string    `json:"account_number" bson:"account_number,omitempty"`
	AccountName        string    `json:"account_name" bson:"account_name,omitempty"`
	AccountType        string    `json:"account_type" bson:"account_type,omitempty"`
	BankCode           string    `json:"bank_code" bson:"bank_code,omitempty"`
	ExpiredDate        string    `json:"expire_date" bson:"expire_date,omitempty"`
	CreatedAt          time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" bson:"updated_at"`
	MerchantId         string    `json:"merchant_id" bson:"merchant_id"`
	MerchantCode       string    `json:"merchant_code" bson:"merchant_code"`
	MapId              string    `json:"map_id" bson:"map_id"`
	MapType            string    `json:"map_type" bson:"map_type"`
	CloseReason        string    `json:"close_reason,omitempty" bson:"close_reason,omitempty"`
	ReOpenTime         time.Time `json:"reopen_time" bson:"reopen_time,omitempty"`
	LimitAmount        int64     `json:"limit_amount" bson:"limit_amount,omitempty"`
	ClosedBy           string    `json:"closed_by,omitempty" bson:"closed_by,omitempty"`
	IsAutoIncrement    bool      `json:"is_auto_increment,omitempty" bson:"-"`
	MerchantIdentifier string    `json:"merchant_identifier,omitempty" bson:"merchant_identifier,omitempty"`

	ManyAccountTypeLimit int64 `json:"account_type_many_limit,omitempty" bson:"account_type_many_limit"`
	MinAmount            int64 `json:"min_amount" bson:"min_amount"`
	MaxAmount            int64 `json:"max_amount" bson:"max_amount"`
	EqualAmount          int64 `json:"equal_amount" bson:"equal_amount"`
}

func (va *VirtualAccounts) ConvertToProto() *order_system.DetailVA {
	vaDto := &order_system.DetailVA{
		AccountNumber:        va.AccountNunmber,
		AccountName:          va.AccountName,
		Status:               va.Status,
		Balance:              va.Balance,
		ExpiredDate:          va.ExpiredDate,
		MapId:                va.MapId,
		MapType:              va.MapType,
		AccountType:          va.AccountType,
		LimitAmount:          va.LimitAmount,
		Provider:             va.Provider,
		MaxAmount:            va.MaxAmount,
		MinAmount:            va.MinAmount,
		EqualAmount:          va.EqualAmount,
		AccountTypeManyLimit: va.ManyAccountTypeLimit,
	}

	createdAt, err := ptypes.TimestampProto(va.CreatedAt)
	if err == nil {
		vaDto.StartedAt = createdAt
	}

	updatedAt, err := ptypes.TimestampProto(va.UpdatedAt)
	if err == nil {
		vaDto.UpdatedAt = updatedAt
	}

	return vaDto
}
