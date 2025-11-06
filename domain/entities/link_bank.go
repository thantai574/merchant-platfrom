package entities

type LinkedBankLink struct {
	ID string `json:"id"` // Link id
	//Direct true thẳng đến bank || false NAPAS
	Direct   bool   `json:"direct"` // Lien ket truc tiep hay gian tiep
	BankCode string `json:"bank_code"`
	BankName string `json:"bank_name"`
	BankLogo string `json:"bank_logo"`
	//URLDirectLinkBank Link chuyển mở web ngân hàng
	URLDirectLinkBank string `json:"url_direct_link_bank" bson:"url_direct_link_bank"`
	URLDeposit        string `json:"url_deposit" bson:"url_deposit"`
	BackgroundSmall   string `json:"background_small" bson:"background_small"`
	BackgroundLarge   string `json:"background_large" bson:"background_large"`
	CardNumber        string `json:"card_number"` // Card number encoded
}
