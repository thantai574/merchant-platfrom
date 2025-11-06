package entities

import "orders-system/domain/constants"

// Rút từ tài khoản bank => nạp vào ví

type CashInRequest struct {
	ClientCode   string            `json:"client_code"`
	GPayBankCode string            `json:"gpay_bank_code"`
	TransTime    int64             `json:"trans_time"`
	Data         CashInDataRequest `json:"data,omitempty"`
	IpAddress    string            `json:"ip_address"`
	Signature    string            `json:"signature"`
}

type CashInDataRequest struct {
	Amount            int64  `json:"amount"`
	GpayTransactionID string `json:"gpay_transaction_id"`
	LinkID            int64  `json:"link_id"`
	Channel           string `json:"channel"`
	GPayUserID        string `json:"gpay_user_id"`
	Description       string `json:"description"`
}

type CashInResponse struct {
	ErrorCode constants.BankGwStatus `json:"error_code"`
	Data      CashInDataResponse     `json:"data"`
	Message   string                 `json:"message"`
	Signature string                 `json:"signature"`
}

type CashInDataResponse struct {
	BankTraceID string `json:"bank_trace_id,omitempty"`
	LinkID      int64  `json:"link_id,omitempty"`
}
