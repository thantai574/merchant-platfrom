package value_objects

type OrderPaymentMerchantStatusDetail struct {
	OrderId  string                     `json:"gpay_transaction_id"`
	Status   OrderPaymentMerchantStatus `json:"status"`
	SuceedAt string                     `json:"suceed_at"`
}

type OrderPaymentMerchantStatus string

func (status OrderPaymentMerchantStatus) IsSuccess() bool {
	if len(status) > 2 && status[0:1] == "2" {
		return true
	}
	return false
}

type RedirectUrl struct {
	Response struct {
		RedirectUrl string `json:"redirect_url"`
	} `json:"response"`
}
