package entities

import "orders-system/domain/constants"

type UnLinkRequest struct {
	ClientCode   string            `json:"client_code"`
	GPayBankCode string            `json:"gpay_bank_code"`
	TransTime    int64             `json:"trans_time"`
	Data         UnLinkRequestData `json:"data"`
	IpAddress    string            `json:"ip_address"`
	Signature    string            `json:"signature"`
}

type UnLinkRequestData struct {
	LinkID      int64  `json:"link_id"`
	Channel     string `json:"channel"`
	Description string `json:"description"`
}

type UnLinkResponse struct {
	ErrorCode constants.BankGwStatus `json:"error_code"`
	Message   string                 `json:"message"`
	Signature string                 `json:"signature"`
}
