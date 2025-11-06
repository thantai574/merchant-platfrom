package entities

type LixiAmount struct {
	ID          string `json:"id" bson:"_id"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
	Status      string `json:"status"`
}
