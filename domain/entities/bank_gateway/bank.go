package entities

import "orders-system/domain/constants"

type Bank struct {
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
	BaseUrl           string `json:"base_url"`
	UrlDirectLinkBank string `json:"url_direct_link_bank"`
	UrlDeposit        string `json:"url_deposit"`
	BackgroundLarge   string `json:"background_large"`
	BackgroundSmall   string `json:"background_small"`
	MinFeeCharge      int64  `json:"min_fee_charge"`
	IbftStatus        string `json:"ibft_status"`
	SortOrder         int64  `json:"sort_order"`
}

type BankRes struct {
	ErrorCode constants.BankGwStatus `json:"error_code"`
	Data      []Bank                 `json:"data,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Signature string                 `json:"signature"`
}
