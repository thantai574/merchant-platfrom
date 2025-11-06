package entities

import "orders-system/domain/constants"

type CheckByPassOTPDataReq struct {
	Amount       string `json:"amount"`
	ApiName      string `json:"api_name"`
	GpayBankCode string `json:"gpay_bank_code"`
	GpayUserId   string `json:"gpay_user_id"`
}

type CheckByPassOTPDataRes struct {
	ErrorCode constants.BankGwStatus `json:"error_code"`
	Data      interface{}            `json:"data,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Signature string                 `json:"signature"`
}
