package entities

import (
	"orders-system/domain/constants"
)

//  Rút tiền về tài khoản ngân hàng

type CashOutRequest struct {
	ClientCode   string             `json:"client_code"`
	GPayBankCode string             `json:"gpay_bank_code"`
	TransTime    int64              `json:"trans_time"`
	Data         CashOutRequestData `json:"data"`
	IpAddress    string             `json:"ip_address"`
	Signature    string             `json:"signature"`
}

type CashOutRequestData struct {
	Amount            int64  `json:"amount"`
	GpayTransactionID string `json:"gpay_transaction_id"`
	LinkID            int64  `json:"link_id"`
	Channel           string `json:"channel"`
	GpayUserID        string `json:"gpay_user_id"`
	Description       string `json:"description"`
}

type CashOutResponse struct {
	ErrorCode constants.BankGwStatus `json:"error_code"`
	Data      CashOutDataResponse    `json:"data,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Signature string                 `json:"signature"`
}

type CashOutDataResponse struct {
	BankTraceId string `json:"bank_trace_id,omitempty"`
}
