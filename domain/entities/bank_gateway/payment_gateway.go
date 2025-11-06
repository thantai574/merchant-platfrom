package entities

import "orders-system/domain/constants"

type PGInitTransReq struct {
	ClientCode   string      `json:"client_code"`
	GPayBankCode string      `json:"gpay_bank_code"`
	TransTime    int64       `json:"trans_time"`
	Data         interface{} `json:"data"`
	IpAddress    string      `json:"ip_address"`
	Signature    string      `json:"signature"`
}

type PGInitTransReqData struct {
	GPayTransactionID string `json:"gpay_transaction_id"`
	Description       string `json:"description"`
	Amount            string `json:"amount"`
	MerchantCode      string `json:"merchant_code"`
	RedirectURL       string `json:"redirect_url"`
}

type PGInitTransRes struct {
	ErrorCode constants.BankGwStatus `json:"error_code"`
	Data      PGInitTransResData     `json:"data,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Signature string                 `json:"signature"`
}

type PGInitTransResData struct {
	BankTraceID string `json:"bank_trace_id"`
	RedirectURL string `json:"redirect_url"`
}

type PGInitOrderRes struct {
	ErrorCode constants.BankGwStatus `json:"error_code"`
	Data      interface{}            `json:"data,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Signature string                 `json:"signature"`
}
