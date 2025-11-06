package entities

import "orders-system/domain/constants"

type LinkInfoRequest struct {
	ClientCode   string              `json:"client_code"`
	GPayBankCode string              `json:"gpay_bank_code"`
	TransTime    int64               `json:"trans_time"`
	Data         LinkInfoRequestData `json:"data"`
	IpAddress    string              `json:"ip_address"`
	Signature    string              `json:"signature"`
}

type LinkInfoRequestData struct {
	LinkID      int64  `json:"link_id"`
	Description string `json:"description"`
}

type LinkInfoResponse struct {
	ErrorCode constants.BankGwStatus `json:"error_code"`
	Data      DataLinkInfoResponse   `json:"data,omitempty"`
	Message   string                 `json:"message"`
	Signature string                 `json:"signature"`
}

type DataLinkInfoResponse struct {
	LinkID        int64  `json:"link_id"`
	CardNumber    string `json:"card_number,omitempty"`    // 6 đầu 4 cuối
	AccountNumber string `json:"account_number,omitempty"` //số tk bị ẩn
	FullName      string `json:"fullName"`
	GpayUserID    string `json:"gpay_user_id"`
	PhoneNumber   string `json:"phone_number"`
	Provider      string `json:"provider"`                 // nha cung cap the
	BankCode      string `json:"bank_code"`                // ma ngan hang
	GpayBankCode  string `json:"gpay_bank_code"`           // ma bank Code Gpay
	Status        string `json:"status"`                   // trạng thái lk thẻ
	FundingMethod string `json:"funding_method,omitempty"` // loại thẻ quốc tế
	Brand         string `json:"brand,omitempty"`          // VISA/ MASTER
}
