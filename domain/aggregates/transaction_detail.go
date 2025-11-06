package aggregates

import "orders-system/domain/entities"

type TransactionDetail struct {
	entities.Transaction
	Payer entities.User `json:"payer"`
	Payee entities.User `json:"payee"`
}
