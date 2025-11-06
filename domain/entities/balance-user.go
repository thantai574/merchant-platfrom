package entities

type Balances struct {
	Id              string   `json:"_id" bson:"_id"`
	AmountAvailable int64    `json:"available" bson:"amount"`
	AmountFreeze    int64    `json:"freeze" bson:"amount_freeze"`
	Type            string   `json:"id" bson:"type"`
	UserId          string   `json:"user_id" bson:"user_id"`
	FreezeIds       []string `json:"freeze_ids" bson:"freeze_ids"`
	Currency        string   `json:"currency" bson:"currency"`
}
