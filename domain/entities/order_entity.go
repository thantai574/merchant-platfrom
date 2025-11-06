package entities

import (
	"orders-system/proto/order_system"
	"orders-system/utils/helpers"
	"time"

	"github.com/golang/protobuf/ptypes"
)

type OrderEntity struct {
	OrderID              string       `json:"order_id" bson:"order_id,omitempty"`
	ServiceID            string       `bson:"service_id,omitempty"`
	RefID                string       `bson:"ref_id,omitempty"`
	UserID               string       `bson:"user_id,omitempty"`
	SubscribeMerchantID  string       `json:"subscribe_merchant_id" bson:"subscribe_merchant_id,omitempty"`
	TransactionID        string       `json:"transaction_id" bson:"transaction_id,omitempty"`
	ServiceType          string       `json:"service_type" bson:"service_type,omitempty"`
	ServiceCode          string       `json:"service_code" bson:"service_code,omitempty"` //  app_id Trans
	OrderType            string       `bson:"order_type,omitempty"`
	SubOrderType         string       `bson:"sub_order_type,omitempty"`
	Amount               int64        `json:"amount" bson:"amount,omitempty"`
	Quantity             int64        `json:"quantity" bson:"quantity,omitempty"`
	SourceOfFund         string       `json:"source_of_fund" bson:"source_of_fund,omitempty"`
	Status               EntityStatus `json:"status" bson:"status,omitempty"`
	VoucherCode          string       `json:"voucher_code" bson:"voucher_code,omitempty"`
	DeviceID             string       `json:"device_id" bson:"device_id,omitempty"`
	MerchantFeeAmount    int64        `json:"merchant_fee_amount" bson:"merchant_fee_amount,omitempty"`
	CreatedAt            time.Time    `json:"created_at" bson:"created_at"`
	UpdatedAt            time.Time    `json:"updated_at" bson:"updated_at,omitempty"`
	DeletedAt            time.Time    `json:"deleted_at" bson:"deleted_at,omitempty"`
	BankCode             string       `json:"bank_code" bson:"bank_code,omitempty"`
	GPayBankCode         string       `json:"g_pay_bank_code" bson:"g_pay_bank_code,omitempty"`
	PhoneTopUp           string       `json:"phone_top_up" bson:"phone_top_up,omitempty"`
	ToUserID             string       `json:"to_user_id" bson:"to_user_id,omitempty"`
	LuckyMoneyID         string       `json:"lucky_money_id" bson:"lucky_money_id,omitempty"`
	ExpiredAt            time.Time    `json:"expired_at"   bson:"expired_at,omitempty"`
	IsExpired            bool         `json:"is_expired"   bson:"is_expired,omitempty"`
	MerchantCode         string       `json:"merchant_code"   bson:"merchant_code,omitempty"`
	MerchantCategoryCode string       `json:"merchant_category_code"   bson:"merchant_category_code,omitempty"`
	MerchantCategoryType string       `json:"merchant_category_type"   bson:"merchant_category_type,omitempty"`
	MerchantTypeWallet   string       `json:"merchant_type_wallet"   bson:"merchant_type_wallet,omitempty"`
	FailReason           string       `json:"fail_reason"   bson:"fail_reason,omitempty"`
	InternalErr          string       `json:"internal_err" bson:"internal_err,omitempty"`

	CardNumber string    `json:"card_number" bson:"card_number,omitempty"`
	AccountNo  string    `json:"account_no" bson:"account_no,omitempty"`
	SucceedAt  time.Time `json:"succeed_at" bson:"succeed_at,omitempty"`

	Napas                     bool     `json:"napas" bson:"napas,omitempty"`
	BankTransactionId         string   `json:"bank_transaction_id" bson:"bank_transaction_id,omitempty"`
	AmountMerchantFee         int64    `json:"amount_merchant_fee" bson:"amount_merchant_fee,omitempty"`
	AmountMerchantFeeGpayTmp  int64    `json:"amount_merchant_fee_gpay_tmp" bson:"amount_merchant_fee_gpay_tmp,omitempty"`
	FixedFeeAmount            int64    `json:"fixed_fee_amount" bson:"fixed_fee_amount,omitempty"`
	RateFeeAmount             int64    `json:"rate_fee_amount" bson:"rate_fee_amount,omitempty"`
	MerchantFeeMethod         string   `json:"merchant_fee_method" bson:"merchant_fee_method,omitempty"`
	PaymentOrderId            string   `json:"payment_order_id" bson:"payment_order_id,omitempty"`
	Description               string   `json:"description" bson:"description,omitempty"`
	TransactionIDs            []string `json:"transaction_ids" bson:"transaction_ids,omitempty"`
	IsRetrievedOrder          bool     `json:"is_retrieved_order" bson:"is_retrieved_order,omitempty"`
	RefundTransactionId       string   `json:"refund_transaction_id" bson:"refund_transaction_id,omitempty"`
	RefundSourceTransactionId string   `json:"refund_source_transaction_id" bson:"refund_source_transaction_id,omitempty"`
	RefundSourceOrderId       string   `json:"refund_source_order_id" bson:"refund_source_order_id,omitempty"`
	RefundType                string   `json:"refund_type" bson:"refund_type,omitempty"`
	ExpireTime                int64    `json:"expire_time,omitempty" bson:"expire_time,omitempty"`
	CustomerId                string   `json:"customer_id,omitempty" bson:"customer_id,omitempty"`
	Metadata                  string   `json:"metadata,omitempty" bson:"metadata,omitempty"`

	OrderCardTelco       string `json:"order_card_telco,omitempty" bson:"order_card_telco,omitempty"`
	OrderBillServiceCode string `json:"order_bill_service_code,omitempty" bson:"order_bill_service_code,omitempty"`
	OrderBillCustomerRef string `json:"order_bill_customer_ref,omitempty" bson:"order_bill_customer_ref,omitempty"`
	OrderBillAreaCode    string `json:"order_bill_area_code,omitempty" bson:"order_bill_area_code,omitempty"`

	BankReceived string `json:"bank_received,omitempty" bson:"bank_received,omitempty"`
	AccountName  string `json:"account_name" bson:"account_name"`
}

func (o *OrderEntity) ConvertToProto() (pr *order_system.OrderEntity) {

	pr = &order_system.OrderEntity{
		OrderID:             o.OrderID,
		ServiceID:           o.ServiceID,
		UserID:              o.UserID,
		TransactionID:       o.TransactionID,
		OrderType:           o.OrderType,
		SubOrderType:        o.SubOrderType,
		SourceOfFund:        o.SourceOfFund,
		Status:              o.Status.StatusOrderProto(),
		VoucherCode:         o.VoucherCode,
		DeviceID:            o.DeviceID,
		MerchantFeeAmount:   o.MerchantFeeAmount,
		RefID:               o.RefID,
		Amount:              o.Amount,
		IsExpired:           o.IsExpired,
		SubcriberMerchantId: o.SubscribeMerchantID,
		RefundId:            o.RefundTransactionId,
		Description:         o.Description,
		CustomerId:          o.CustomerId,
		Metadata:            o.Metadata,
		GpayBankCode:        o.GPayBankCode,
	}

	created_at, err := ptypes.TimestampProto(o.CreatedAt)
	if err == nil {
		pr.CreatedAt = created_at
	}

	updated_at, err := ptypes.TimestampProto(o.UpdatedAt)
	if err == nil {
		pr.UpdatedAt = updated_at
	}

	deleted_at, err := ptypes.TimestampProto(o.DeletedAt)
	if err == nil {
		pr.DeletedAt = deleted_at
	}

	expired_at, err := ptypes.TimestampProto(o.ExpiredAt)
	if err == nil {
		pr.ExpiredAt = expired_at
	}

	succeed_at, err := ptypes.TimestampProto(o.SucceedAt)
	if err == nil {
		pr.SucceedAt = succeed_at
	}

	return
}

func (o *OrderEntity) ProtoToEntity(in *order_system.OrderRequest) (pr *order_system.OrderEntity) {
	if in.RefID != "" {
		o.RefID = in.RefID
	}
	if in.BankTransactionId != "" {
		o.BankTransactionId = in.BankTransactionId
	}
	if in.Status != order_system.OrderStatus_UNKNOWN {
		o.Status = EntityStatus(in.Status)
	}
	if in.BankCode != "" {
		o.BankCode = in.BankCode
	}
	if in.GPayBankCode != "" {
		o.GPayBankCode = in.GPayBankCode
	}
	if in.Napas {
		o.Napas = in.Napas
	}
	if in.CustomerId != "" {
		o.CustomerId = in.CustomerId
	}
	if in.Metadata != "" {
		o.Metadata = in.Metadata
	}
	if in.FailReason != "" {
		o.FailReason = in.FailReason
	}
	if in.InternalErr != "" {
		o.InternalErr = in.InternalErr
	}
	o.UpdatedAt = helpers.GetCurrentTime()
	return
}
