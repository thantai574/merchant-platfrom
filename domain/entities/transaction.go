package entities

import "time"

type Transaction struct {
	Id string `json:"id" bson:"_id,omitempty"`

	TransactionId string `json:"transaction_id" bson:"transaction_id"`
	AppId         string `json:"app_id" bson:"app_id"`

	Currency string `json:"currency" bson:"currency"`
	Message  string `json:"message" bson:"message"`

	State  string `json:"state" bson:"state"`
	Status string `json:"status" bson:"status"`

	ServiceType        string `json:"service_type,omitempty" bson:"service_type"`
	TransactionType    string `json:"transaction_type,omitempty" bson:"transaction_type"`
	SubTransactionType string `json:"sub_transaction_type,omitempty" bson:"sub_transaction_type"`
	TypeWallet         string `json:"type_wallet,omitempty" bson:"type_wallet"`

	Amount                    uint64 `json:"amount"`
	LastAmount                int64  `json:"last_amount,omitempty" bson:"last_amount"`
	AmountFeeGpay             int64  `json:"amount_fee_gpay,omitempty" bson:"amount_fee_gpay,omitempty"`
	DiscountAmount            int64  `json:"amount_discount,omitempty" bson:"amount_discount"`
	AmountMerchantFee         int64  `json:"amount_merchant_fee,omitempty" bson:"amount_merchant_fee"`
	AmountTransactionCashBack int64  `json:"amount_transaction_cash_back,omitempty" bson:"amount_transaction_cash_back"`

	VoucherCode  string `json:"voucher_code,omitempty" bson:"voucher_code"`
	Source       string `json:"source,omitempty" bson:"source"`
	SourceOfFund string `json:"source_of_fund,omitempty" bson:"source_of_fund"`

	MerchantId            string `json:"merchant_id,omitempty" bson:"merchant_id"`
	MerchantTransactionId string `json:"merchant_transaction_id,omitempty" bson:"merchant_transaction_id"`
	TransactionCashback   string `json:"transaction_cashback,omitempty" bson:"transaction_cashback"`
	RefId                 string `json:"ref_id,omitempty" bson:"ref_id,omitempty"`
	OrderId               string `json:"order_id,omitempty" bson:"order_id,omitempty"`

	DeviceId string `json:"device_id,omitempty" bson:"device_id"`
	PayerId  string `json:"payer_id,omitempty" bson:"payer_id"`
	PayeeId  string `json:"payee_id,omitempty" bson:"payee_id"`

	PayerStatusKyc        bool `json:"payer_status_kyc" bson:"payer_status_kyc"`
	PayerStatusLinkedBank bool `json:"payer_status_linked_bank" bson:"payer_status_linked_bank"`
	PayeeStatusKyc        bool `json:"payee_status_kyc" bson:"payee_status_kyc"`
	PayeeStatusLinkedBank bool `json:"payee_status_linked_bank" bson:"payee_status_linked_bank"`

	GpayAccountID string `json:"gpay_account_id,omitempty" bson:"gpay_account_id,omitempty"`

	BankTransactionId string `json:"bank_transaction_id" bson:"bank_transaction_id"`
	Napas             bool   `json:"napas" bson:"napas"`
	BankCode          string `json:"bank_code" bson:"bank_code"`

	IbftType      string `json:"ibft_type" bson:"ibft_type"`
	CardNo        string `json:"card_no" bson:"card_no"`
	AccountNo     string `json:"account_no" bson:"account_no"`
	FailReason    string `json:"fail_reason" bson:"fail_reason"`
	ReturnCode    string `json:"return_code" bson:"return_code"`
	ReturnMessage string `json:"return_message" bson:"return_message"`
	Exception     string `json:"exception" bson:"exception"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
	DeletedAt time.Time `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`

	Type                   string `json:"type" bson:"type"`
	FeeGpay                int64  `json:"fee_gpay" bson:"fee_gpay"`
	AmountAfterAddVoucher  int64  `json:"amount_after_add_voucher,omitempty" bson:"amount_after_add_voucher"`
	AmountBeforeAddVoucher int64  `json:"amount_before_add_voucher,omitempty" bson:"amount_before_add_voucher"`
	InvoiceId              string `bson:"invoice_id,omitempty" json:"invoice_id"`
	TypeTransfer           string `json:"type_transfer,omitempty" bson:"type_transfer"`
	PayerAmountBefore      int64  `json:"payer_amount_before,omitempty" bson:"payer_amount_before"`
	PayeeAmountBefore      int64  `json:"payee_amount_before,omitempty" bson:"payee_amount_before"`
	PayerAmountAfter       int64  `json:"payer_amount_after,omitempty" bson:"payer_amount_after"`
	PayeeAmountAfter       int64  `json:"payee_amount_after,omitempty" bson:"payee_amount_after"`
	UnUseWallet            bool   `json:"un_use_wallet,omitempty" bson:"un_use_wallet"`
	PartnerCode            string `json:"partner_code,omitempty" bson:"partner_code"`
	WalletIdSrc            string `json:"wallet_id_src,omitempty" bson:"wallet_id_src"`
	WalletIdDist           string `json:"wallet_id_dist,omitempty" bson:"wallet_id_dist"`
	V                      int    `json:"v,omitempty" bson:"_v"`
	PaymentType            string `json:"payment_type,omitempty"`

	ProviderMerchantId      string `json:"provider_merchant_id,omitempty" bson:"provider_merchant_id,omitempty"`
	TransactionSourceRefund string `json:"transaction_source_refund,omitempty" bson:"transaction_source_refund,omitempty"`
	UserReceiveRefund       string `json:"user_receive_refund,omitempty" bson:"user_receive_refund,omitempty"`
	MerchantCode            string `json:"merchant_code,omitempty" bson:"merchant_code,omitempty"`
	RedirectUrl             string `json:"redirect_url,omitempty" bson:"redirect_url,omitempty"`
}
