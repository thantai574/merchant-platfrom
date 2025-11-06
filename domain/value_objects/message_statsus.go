package value_objects

type MessagesStatus struct {
	IsRead    bool   `json:"is_read" bson:"is_read"`
	MessageId string `json:"message_id" bson:"message_id"`
	UserId    string `json:"user_id" bson:"user_id"`
}

type URI struct {
	Type string `json:"type"`
	Data string `json:"data"`
}
