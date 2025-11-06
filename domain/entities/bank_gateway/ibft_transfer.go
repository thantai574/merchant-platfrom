package entities

import "orders-system/domain/constants"

type IBFTTransferRequest struct {
	ClientCode   string              `json:"client_code"`
	GPayBankCode string              `json:"gpay_bank_code"`
	TransTime    int64               `json:"trans_time"`
	Data         IBFTTransferReqData `json:"data"`
	IpAddress    string              `json:"ip_address"`
	Signature    string              `json:"signature"`
}

type IBFTTransferReqData struct {
	AccountNumber     string `json:"account_number"`
	CardNumber        string `json:"card_number"`
	GpayUserID        string `json:"gpay_user_id"`
	GpayTransactionID string `json:"gpay_transaction_id"`
	Amount            int64  `json:"amount"`
	IBFTCode          string `json:"bank_bin"`
	Description       string `json:"description"`
	BankName          string `json:"bank_name"`
}

type IBFTTransferResponse struct {
	ErrorCode constants.BankGwStatus   `json:"error_code"`
	Data      IBFTTransferResponseData `json:"data"`
	Message   string                   `json:"message"`
	Signature string                   `json:"signature"`
}

type IBFTTransferResponseData struct {
	BankTraceId string `json:"bank_trace_id,omitempty"`
}
