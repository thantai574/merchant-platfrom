package entities

import "orders-system/domain/constants"

type CheckBankBinResponse struct {
	ErrorCode constants.BankGwStatus   `json:"error_code"`
	Data      CheckBankBinDataResponse `json:"data,omitempty"`
	Message   string                   `json:"message,omitempty"`
	Signature string                   `json:"signature"`
}

type CheckBankBinDataResponse struct {
	Data []struct {
		CardType CardType `json:"card_type"`
		BankName string   `json:"bank_name"`
	} `json:"data"`
}

type CardType string

func (cardType CardType) IsCredit() bool {
	if cardType == "CREDIT" {
		return true
	}
	return false
}

func (cardType CardType) IsDebit() bool {
	if cardType == "DEBIT" {
		return true
	}
	return false
}
