package request_params

type GetMerchantConfigReq struct {
	MerchantId   string `json:"merchant_id"`
	ServiceType  string `json:"service_type"`
	TransType    string `json:"trans_type"`
	SubTransType string `json:"sub_trans_type"`
}

type GetRefundConfigReq struct {
	MerchantId   string `json:"merchant_id"`
	ServiceType  string `json:"service_type"`
	TransType    string `json:"trans_type"`
	SubTransType string `json:"sub_trans_type"`
	SourceOfFund string `json:"source_of_fund"`
}
