package entities

type IBFTGpayBalanceCheckRequest struct {
	ClientCode   string                   `json:"clientCode"`
	GPayBankCode string                   `json:"gpayBankCode"`
	TransTime    int64                    `json:"transTime"`
	Data         IBFTGpayBalanceCheckData `json:"data"`
	Signature    string                   `json:"signature"`
}

type IBFTGpayBalanceCheckData struct {
	AccountNumber string `json:"accountNumber"`
	Channel       string `json:"channel"`
	Description   string `json:"description"`
}

type IBFTGpayBalanceCheckResponse struct {
	ErrorCode string                      `json:"errorCode"`
	Data      IBFTGPayBalanceResponseData `json:"data,omitempty"`
	Message   string                      `json:"message,omitempty"`
	Signature string                      `json:"signature"`
}

type IBFTGPayBalanceResponseData struct {
	Balance       int64  `json:"balance,omitempty"`
	AccountNumber string `json:"accountNumber,omitempty"` // số tk
	AccountName   string `json:"acc,omitempty"`           // tên tk
}
