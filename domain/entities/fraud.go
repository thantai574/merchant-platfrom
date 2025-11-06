package entities

type Fraud struct {
	Id         string      `json:"id"`
	FullName   string      `json:"full_name"`
	CardNumber string      `json:"cardNumber"`
	Status     StatusFraud `json:"status"`
	Reason     string      `json:"reason"`
	ReasonCode string      `json:"reason_code"`
}

type StatusFraud string

func (status StatusFraud) IsFail() bool {
	if status == "FAIL" {
		return true
	}
	return false
}

func (status StatusFraud) IsSuccess() bool {
	if status == "SUCCESS" {
		return true
	}
	return false
}

func (status StatusFraud) IsProcessing() bool {
	if status == "PROCESSING" {
		return true
	}
	return false
}
