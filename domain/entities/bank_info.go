package entities

type Bank struct {
	BankId    int64  `json:"bank_id" bson:"_id"`
	Logo      string `json:"logo"`
	Name      string `json:"name"`
	ShortName string `json:"bank_code" bson:"short_name"`
	//Direct true thẳng đến bank || false NAPAS
	Direct bool `json:"direct"`
	//URLDirectLinkBank Link chuyển mở web ngân hàng
	URLDirectLinkBank string `json:"url_direct_link_bank" bson:"url_direct_link_bank"`
	URLDeposit        string `json:"url_deposit" bson:"url_deposit"`
	ApiEndpoint       string `json:"api_endpoint" bson:"api_endpoint"`
	//Status ACTIVE || PAUSE
	Status          string `json:"status"`
	IBFTCode        string `json:"ibft_code" bson:"ibft_code"`
	EpayBankNo      string `json:"epay_bank_no" bson:"epay_bank_no"`
	EpayBankAccType string `json:"epay_bank_acc_type" bson:"epay_bank_acc_type"`
	MinFeeCharge    int64  `json:"min_fee_charge"  bson:"min_fee_charge"`
	BackgroundSmall string `json:"background_small" bson:"background_small"`
	BackgroundLarge string `json:"background_large" bson:"background_large"`
	Type            string `json:"type"`
}
