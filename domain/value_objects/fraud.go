package value_objects

type FraudRequest struct {
	CardNumber string `json:"cardNumber"`
}

type FraudTransRequest struct {
	TransactionId string      `json:"transactionId"`
	FraudDto      interface{} `json:"fraudDto"`
}
