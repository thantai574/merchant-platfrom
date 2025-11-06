package configs

import (
	"github.com/spf13/viper"
)

type Config struct {
	Prefix            string            `json:"prefix" mapstructure:"prefix"`
	Port              string            `json:"port" mapstructure:"port"`
	ENV               string            `json:"env" mapstructure:"env"`
	Job               bool              `json:"job" mapstructure:"job"`
	MaxPoolSize       int               `json:"max_pool_size" mapstructure:"max_pool_size"`
	MongoURI          string            `json:"mongo_uri" mapstructure:"mongo_uri"`
	KafkaConfig       Kafka             `json:"kafka_config"  mapstructure:"kafka_config"`
	Services          ServicesConfig    `json:"services" mapstructure:"services"`
	QueueUri          string            `json:"queue_uri" mapstructure:"queue_uri"`
	Redis             RedisConfig       `json:"redis"  mapstructure:"redis"`
	LinkedBank        LinkedBank        `json:"linked_bank" mapstructure:"linked_bank"`
	MQTTMobileUri     string            `json:"mqtt_uri" mapstructure:"mqtt_uri"`
	MQTTInternalUri   MQTTInternalUri   `json:"mqtt_internal_uri" mapstructure:"mqtt_internal_uri"`
	BankGwURI         string            `json:"bank_gw_uri" mapstructure:"bank_gw_uri"`
	FirebaseKey       string            `json:"firebase_key" mapstructure:"firebase_key"`
	TelegramChannelId TelegramChannelId `json:"telegram_channel_id" mapstructure:"telegram_channel_id"`
	ExpireTime        int64             `json:"expire_time" mapstructure:"expire_time"`
	FraudUri          string            `json:"fraud_uri" mapstructure:"fraud_uri"`
	MerchantCallBack  MerchantCallBack  `json:"merchant_callback" mapstructure:"merchant_callback"`
	VaMsb             VaMsb             `json:"va_msb" mapstructure:"va_msb"`
}

type VaMsb struct {
	GpayIdentifier string `json:"gpay_identifier" mapstructure:"gpay_identifier"`
}

type MQTTInternalUri struct {
	Uri      string `json:"uri" mapstructure:"uri"`
	Username string `json:"username" mapstructure:"username"`
	Password string `json:"password" mapstructure:"password"`
	Prefix   string `json:"prefix" mapstructure:"prefix"`
}

type TelegramChannelId struct {
	QR           int64 `json:"qr" mapstructure:"qr"`
	Fraud        int64 `json:"fraud" mapstructure:"fraud"`
	FundTransfer int64 `json:"fund_transfer" mapstructure:"fund_transfer"`
	Refund       int64 `json:"refund" mapstructure:"refund"`
	Va           int64 `json:"va" mapstructure:"va"`
}

type MerchantCallBack struct {
	Uri         string `json:"uri" mapstructure:"uri"`
	SecretKey   string `json:"secret_key" mapstructure:"secret_key"`
	SecretValue string `json:"secret_value" mapstructure:"secret_value"`
}

type LinkedBank struct {
	URI          string `json:"uri" mapstructure:"uri"`
	DBName       string `json:"db_name" mapstructure:"db_name"`
	APIEndpoint  string `json:"api_endpoint" mapstructure:"api_endpoint"`
	VccbEndpoint string `json:"vccb_endpoint" mapstructure:"vccb_endpoint"`
}

type Kafka struct {
	Zookeepers       string   `json:"zookeepers" mapstructure:"zookeepers"`
	Brokers          string   `json:"brokers" mapstructure:"brokers"`
	KafkaSubscribers []string `mapstructure:"kafka_subscribers"`
	Partitions       int      `mapstructure:"partitions"`
	Replicas         int      `mapstructure:"replicas"`
	ReturnDuration   int      `mapstructure:"return_duration"`
}

type RedisConfig struct {
	Address  string `json:"address" mapstructure:"address"`
	Password string `json:"password" mapstructure:"password"`
	DB       int    `json:"db" mapstructure:"db"`
}
type ServicesConfig struct {
	ServiceCore        string `json:"service_core" mapstructure:"service_core"`
	ServiceUser        string `json:"service_user" mapstructure:"service_user"`
	ServicePromotion   string `json:"service_promotion" mapstructure:"service_promotion"`
	ServiceTransaction string `json:"service_transaction" mapstructure:"service_transaction"`
	ServiceMerchantFee string `json:"service_merchant_fee" mapstructure:"service_merchant_fee"`
	ServiceCard        string `json:"service_card" mapstructure:"service_card"`
}

func LoadConfig() (*Config, error) {
	viper.AddConfigPath("./")
	viper.SetConfigType("json")
	viper.SetConfigName("config.json")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	result := &Config{}
	err = viper.Unmarshal(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// LoadTestConfig load config for running tests
func LoadTestConfig(configPath string) (*Config, error) {
	viper.AddConfigPath(configPath)
	viper.SetConfigType("json")
	viper.SetConfigName("config_test.json")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	result := &Config{}
	err = viper.Unmarshal(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
