package entities

import "time"

type Devices struct {
	Id                    string    `json:"_id" bson:"_id"`
	MessageTopics         []string  `json:"message_topics" bson:"message_topics"`
	UserId                string    `json:"user_id" bson:"user_id"`
	DeviceName            string    `json:"device_name" bson:"device_name"`
	DeviceOs              string    `json:"device_os" bson:"device_os"`
	Status                string    `json:"status" bson:"status"`
	CreatedAt             time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" bson:"updated_at"`
	DeviceToken           string    `json:"device_token" bson:"device_token"`
	DeviceTokenHash       string    `json:"device_token_hash" bson:"device_token_hash"`
	AccessTokenSalt       string    `json:"access_token_salt" bson:"access_token_salt"`
	RefreshToken          string    `json:"refresh_token" bson:"refresh_token"`
	RefreshTokenCreatedAt time.Time `json:"refresh_token_created_at" bson:"refresh_token_created_at"`
}
