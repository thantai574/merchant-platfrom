package entities

import (
	"orders-system/domain/constants"
)

type ActionVARequest struct {
	ClientCode   string      `json:"client_code"`
	GPayBankCode string      `json:"gpay_bank_code"`
	TransTime    int64       `json:"trans_time"`
	Data         interface{} `json:"data"`
	IpAddress    string      `json:"ip_address"`
	Signature    string      `json:"signature"`
}

type CreateVARequestData struct {
	AccountName     string `json:"account_name"`
	AccountType     string `json:"account_type,omitempty"`
	AccountNumber   string `json:"account_number,omitempty"`
	ReferenceNumber string `json:"reference_number,omitempty"`
	Status          string `json:"status,omitempty"`

	MaxAmount   string `json:"max_amount,omitempty"`
	MinAmount   string `json:"min_amount,omitempty"`
	EqualAmount string `json:"equal_amount,omitempty"`
}

type ActionVAResponse struct {
	ErrorCode constants.BankGwStatus `json:"error_code"`
	Data      DataVADetailResponse   `json:"data,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Signature string                 `json:"signature"`
}

type DataVADetailResponse struct {
	AccountNumber     string `json:"account_number"`
	AccountName       string `json:"account_name"`
	Currency          string `json:"currency"`
	GpayAccountNumber string `json:"gpay_account_number"`
	Status            string `json:"status"` // Trạng thái tài khoản  (OPEN-CLOSE-FREEZE)
	MakerId           string `json:"maker_id"`
	CheckerId         string `json:"checker_id"`
	ExpireDate        string `json:"expire_date"`
	AccountType       string `json:"account_type"`
}

type UpdateVARequestData struct {
	AccountNumber string `json:"account_number"`
	AccountName   string `json:"account_name"`
	Status        string `json:"status,omitempty"`
}

type CloseVARequestData struct {
	AccountNumber string `json:"account_number"`
	Status        string `json:"status,omitempty"`
}

type DetailVARequestData struct {
	AccountNumber string `json:"account_number"`
}

type ReOpenVARequestData struct {
	AccountNumber string `json:"account_number"`
	Status        string `json:"status,omitempty"`
}
