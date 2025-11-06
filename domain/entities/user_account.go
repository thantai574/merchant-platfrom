package entities

type UserAccount struct {
	AccountId           string   `bson:"account_id"`
	AmountAvailable     int64    `bson:"amount_available"`
	TypeWallet          string   `bson:"type_wallet"`
	UserId              string   `bson:"user_id"`
	PendingTransactions []string `bson:"pending_transactions"`
}
