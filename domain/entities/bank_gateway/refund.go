package entities

import "orders-system/domain/constants"

type RefundReq struct {
	ClientCode   string        `json:"client_code"`
	GPayBankCode string        `json:"gpay_bank_code"`
	TransTime    int64         `json:"trans_time"`
	Data         RefundReqData `json:"data"`
	IpAddress    string        `json:"ip_address"`
	Signature    string        `json:"signature"`
}

type RefundReqData struct {
	GPayTransactionID string `json:"gpay_transaction_id"`
	Amount            string `json:"amount"`
}

type RefundRes struct {
	ErrorCode constants.BankGwStatus `json:"error_code"`
	Data      RefundResData          `json:"data,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Signature string                 `json:"signature"`
}

type RefundResData struct {
	BankTransactionId string `json:"bank_transaction_id"`
}
