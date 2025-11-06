package entities

import "orders-system/domain/constants"

type VerifyOTPRequest struct {
	ClientCode   string               `json:"client_code"`
	GPayBankCode string               `json:"gpay_bank_code"`
	TransTime    int64                `json:"trans_time"`
	Data         VerifyOTPRequestData `json:"data"`
	IpAddress    string               `json:"ip_address"`
	Signature    string               `json:"signature"`
}

type VerifyOTPRequestData struct {
	RefBankTraceID string `json:"ref_bank_trace_id"` // mã Trace của Bank forward cho bankGateway
	// Bắt buộc với các gd cashin (nạp  ví)
	LinkID            int64  `json:"link_id"` // Bắt buộc với các yêu cầu liên kết thẻ
	OTP               string `json:"otp"`
	Channel           string `json:"channel" enums:"MOBILE || WEB || POS || DESKTOP || SMS "` // kênh
	Description       string `json:"description"`
	GpayTransactionId string `json:"gpay_transaction_id"`
}

type VerifyOTPResponse struct {
	ErrorCode constants.BankGwStatus `json:"error_code"`
	Data      VerifyDetail           `json:"data,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Signature string                 `json:"signature"`
}

type VerifyDetail struct {
	BankTraceID string `json:"bank_trace_id,omitempty"`
}
