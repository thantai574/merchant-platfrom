package entities

// @todo Kiểm tra số dư tài khoản đảm bảo của Gpay tại Bank tương ứng
type BalanceCheckRequest struct {
	ClientCode   string                  `json:"clientCode"`
	GPayBankCode string                  `json:"gpayBankCode"`
	TransTime    int64                   `json:"transTime"`
	Data         BalanceCheckRequestData `json:"data"`
	Signature    string                  `json:"signature"`
}

type BalanceCheckRequestData struct {
	AccountNumber string `json:"accountNumber"` // số tài khoản
	Channel       string `json:"channel"`
	Description   string `json:"description"`
}

type BalanceCheckResponse struct {
	ErrorCode string      `json:"errorCode"`
	Data      DataBalance `json:"data"`
	Message   string      `json:"message"`
	Signature string      `json:"signature"`
}

type DataBalance struct {
	Balance       int64  `json:"balance"`
	AccountNumber string `json:"accountNumber"` // số tài khoản
	AccountName   string `json:"accountName"`   // tên tk
}
