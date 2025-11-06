package entities

import "orders-system/domain/constants"

type IBFTInquiryCheckRequest struct {
	ClientCode   string               `json:"client_code"`
	GPayBankCode string               `json:"gpay_bank_code"`
	TransTime    int64                `json:"trans_time"`
	Data         IBFTInquiryCheckData `json:"data"`
	IpAddress    string               `json:"ip_address"`
	Signature    string               `json:"signature"`
}

type IBFTInquiryCheckData struct {
	AccountNumber string `json:"account_number"`
	CardNumber    string `json:"card_number"`
	BankBin       string `json:"bank_bin"`
	Description   string `json:"description"`
}

type IBFTInquiryCheckResponse struct {
	ErrorCode constants.BankGwStatus `json:"error_code"`
	Data      IBFTCheckResponseData  `json:"data,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Signature string                 `json:"signature"`
}

type IBFTCheckResponseData struct {
	FullName string `json:"full_name,omitempty"`
}
