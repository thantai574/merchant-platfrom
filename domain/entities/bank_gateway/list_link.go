package entities

import "orders-system/domain/constants"

type ListLinkRequest struct {
	ClientCode   string              `json:"client_code"`
	GPayBankCode string              `json:"gpay_bank_code"`
	TransTime    int64               `json:"trans_time"`
	Data         ListLinkRequestData `json:"data"`
	IpAddress    string              `json:"ip_address"`
	Signature    string              `json:"signature"`
}

type ListLinkRequestData struct {
	GpayUserID  string `json:"gpay_user_id"`
	Description string `json:"description"`
}

type ListLinkResponse struct {
	ErrorCode constants.BankGwStatus `json:"err_code"`
	Data      []*ListLinked          `json:"data,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Signature string                 `json:"signature"`
}

type ListLinked struct {
	LinkID          int64  `json:"link_id,omitempty"`
	CardNumber      string `json:"card_number,omitempty"`
	AccountNumber   string `json:"account_number,omitempty"`
	FullName        string `json:"full_name,omitempty"`
	GpayUserID      string `json:"gpay_user_id,omitempty"`
	PhoneNumber     string `json:"phone_number,omitempty"`
	Provider        string `json:"provider,omitempty"`
	BankCode        string `json:"bank_code,omitempty"`
	Status          string `json:"status,omitempty"`
	GpayBankCode    string `json:"gpay_bank_code,omitempty"`
	Logo            string `json:"logo,omitempty"`
	BankPublishName string `json:"bank_publish_name,omitempty"`

	UrlDirectLinkBank string `json:"url_direct_link_bank,omitempty"`
	UrlDeposit        string `json:"url_deposit,omitempty"`
	BackgroundLarge   string `json:"background_large,omitempty"`
	BackgroundSmall   string `json:"background_small,omitempty"`
	FundingMethod     string `json:"funding_method,omitempty"` // loại thẻ quốc tế
}

type ListBankResponse struct {
	ErrorCode string      `json:"errorCode"`
	Data      []*ListBank `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Signature string      `json:"signature"`
}

type ListBank struct {
	Id                int64  `json:"id"`
	BankName          string `json:"bank_name"`
	GpayBankCode      string `json:"gpay_bank_code"`
	Provider          string `json:"provider"`
	ShortNameByNapas  string `json:"short_name_by_napas"`
	ShortNameByGpay   string `json:"short_name_by_gpay"`
	PublishName       string `json:"publish_name"`
	BankBin           string `json:"bank_bin"`
	Status            string `json:"status"`
	CreateTime        int64  `json:"create_time"`
	Logo              string `json:"logo"`
	UrlDirectLinkBank string `json:"url_direct_link_bank,omitempty"`
	IBFTStatus        string `json:"ibft_status,omitempty"`
	UrlDeposit        string `json:"url_deposit"`
	MinFeeCharge      int64  `json:"min_fee_charge"  `
	BackgroundSmall   string `json:"background_small" `
	BackgroundLarge   string `json:"background_large" `
	SortOrder         int64  `json:"sort_order"`
}
