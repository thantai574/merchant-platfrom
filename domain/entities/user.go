package entities

import "time"

type User struct {
	Id                           string    `json:"id" bson:"_id,omitempty"`
	PhoneNumber                  string    `json:"phone_number" bson:"phone_number,omitempty"`
	Email                        string    `json:"email" bson:"email,omitempty"`
	Name                         string    `json:"name"  bson:"name,omitempty"`
	Avatar                       string    `json:"avatar" bson:"avatar,omitempty"`
	Gender                       string    `json:"gender" bson:"gender,omitempty"`
	Status                       string    `json:"status" bson:"status,omitempty"`
	Timezone                     string    `json:"timezone"`
	Language                     string    `json:"language"`
	Title                        string    `json:"title"`
	DateTime                     int64     `json:"date_time" bson:"date_time,omitempty"`
	Currency                     string    `json:"currency" bson:"currency"`
	Role                         string    `json:"role" bson:"role"`
	KYC                          string    `json:"kyc" bson:"kyc" enums:"ACTIVE|PENDING|FAILED|INACTIVE"`
	IdentityNumber               string    `json:"identity_number" bson:"identity_number"`
	BirthDay                     string    `json:"birth_day" bson:"birth_day"`
	Address                      string    `json:"address" bson:"address"`
	MessageTopics                []string  `json:"message_topics" bson:"message_topics"`
	DeviceIdLogin                string    `json:"device_id_login" bson:"device_id_login"`
	LastDeviceIdLogin            string    `json:"last_device_id_login" bson:"last_device_id_login"`
	Password                     string    `json:"-" bson:"password"`
	Long                         float64   `json:"long"`
	Lat                          float64   `json:"lat"`
	OriginDevice                 string    `json:"origin_device" bson:"origin_device"`
	Source                       string    `json:"source"`
	RandomString                 string    `json:"-" bson:"random_string"`
	LastAmountNetPrimaryWallet   int64     `json:"last_amount_net_primary_wallet" bson:"last_amount_net_primary_wallet"`
	LastAmountPrimaryWallet      int64     `json:"last_amount_primary_wallet" bson:"last_amount_primary_wallet"`
	IncrementAmountPrimaryWallet int64     `json:"increment_amount_primary_wallet" bson:"increment_amount_primary_wallet"`
	DecrementAmountPrimaryWallet int64     `json:"decrement_amount_primary_wallet" bson:"decrement_amount_primary_wallet"`
	CreatedAt                    time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt                    time.Time `json:"updated_at" bson:"updated_at"`
	ThreadsHold                  int64     `json:"threads_hold" bson:"threads_hold"`
	HasEverLinkedBank            bool      `json:"has_ever_linked_bank,omitempty" bson:"has_ever_linked_bank,omitempty"`
	GapoId                       string    `json:"gapo_id" bson:"gapo_id"`
	GapoName                     string    `json:"gapo_name" bson:"gapo_name"`
	PinCode                      string    `json:"-"`
	PinCodeExpired               int64     `json:"-"`
	LockTimeExpired              int64     `json:"-"`
	WrongPinCode                 int64     `json:"-"`
}
