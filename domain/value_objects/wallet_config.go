package value_objects

import "time"

type GetMerchantConfigRes struct {
	MerchantIds []string                `json:"merchant_ids" bson:"merchant_ids"`
	Settings    []MerchantConfigSetting `json:"settings" bson:"settings"`
}

type MerchantConfigSetting struct {
	ServiceType   string   `json:"service_type" bson:"service_type"`
	TransType     string   `json:"trans_type" bson:"trans_type"`
	Provider      string   `json:"provider" bson:"provider"`
	Status        string   `json:"status" bson:"status"`
	SubTransTypes []string `json:"sub_trans_types,omitempty" bson:"sub_trans_types,omitempty"`
	VACode        string   `json:"va_code,omitempty" bson:"va_code,omitempty"`
}

type GetRefundConfigRes struct {
	ServiceType  string          `json:"service_type" bson:"service_type"`
	TransType    string          `json:"trans_type" bson:"trans_type"`
	SubTransType []string        `json:"sub_trans_types" bson:"sub_trans_types"`
	MerchantIds  []string        `json:"merchant_ids" bson:"merchant_ids"`
	Settings     []RefundSetting `json:"settings" bson:"settings"`
	CreatedAt    time.Time       `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" bson:"updated_at"`
}

type RefundSetting struct {
	CreateBy        string   `json:"create_by" bson:"create_by"`
	SourceOfFunds   []string `json:"source_of_funds" bson:"source_of_funds"`
	PaymentStatus   string   `json:"payment_status" bson:"payment_status"`     // UNPAID / PAID
	RefundType      string   `json:"refund_type" bson:"refund_type"`           // INDIRECT / API
	ConfirmType     string   `json:"confirm_type" bson:"confirm_type"`         // MANUAL / AUTO
	RefundCondition string   `json:"refund_condition" bson:"refund_condition"` // LT / LE
	RefundValue     int64    `json:"refund_value" bson:"refund_value"`
	Status          string   `json:"status" bson:"status"`
}
