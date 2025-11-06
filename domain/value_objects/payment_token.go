package value_objects

type InitPaymentTokenResponse struct {
	BankTraceId  string `json:"bank_trace_id"`
	GpayBankCode string `json:"gpay_bank_code"`
	OrderId      string `json:"order_id"`
}
