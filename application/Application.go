package application

import (
	"github.com/micro/go-micro/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"orders-system/domain/constants"
	"orders-system/domain/repositories"
	"orders-system/infrastructure/database_mgo"
	"orders-system/infrastructure/database_mgo/devices"
	"orders-system/infrastructure/database_mgo/lixi"
	"orders-system/infrastructure/database_mgo/messages"
	"orders-system/infrastructure/database_mgo/order"
	"orders-system/infrastructure/database_mgo/va_VCCB"
	"orders-system/infrastructure/database_mgo/wallet_config"
	"orders-system/infrastructure/firebase"
	"orders-system/infrastructure/kafka"
	"orders-system/infrastructure/mqtt"
	"orders-system/infrastructure/service/fraud"

	//"orders-system/infrastructure/mqtt"
	mqttCl "github.com/eclipse/paho.mqtt.golang"
	"orders-system/infrastructure/rabbitmq"
	"orders-system/infrastructure/service/bank_service"
	"orders-system/proto/service_card"
	"orders-system/proto/service_merchant_fee"
	"orders-system/proto/service_promotion"
	"orders-system/proto/service_transaction"
	"orders-system/proto/service_user"
	"orders-system/utils/configs"
	"orders-system/utils/gpooling"
	"orders-system/utils/saga"
)

type OrderApplication struct {
	Config                *configs.Config
	Queue                 *rabbitmq.RabbiMQ
	Logger                *zap.Logger
	OrderRepository       repositories.OrderRepository
	TransactionRepository service_transaction.Service
	PromotionRepository   service_promotion.PromotionServicesService
	UserRepository        service_user.Service
	MerchantFeeRepository service_merchant_fee.Service
	ServiceCardRepository service_card.ServiceCardClient
	BankRepository        repositories.BankRepository
	BankServiceRepository repositories.BankServiceRepository
	IPool                 gpooling.IPool
	LogSaga               saga.Store
	LixiRepository        repositories.LixiRepository
	MQTT                  repositories.IMqtt
	KafkaConnection       kafka.Storage
	IMessage              repositories.IMessage
	IDevice               repositories.IDevice
	IFirebase             firebase.IFirebase
	IVA                   repositories.IVA
	IFraud                repositories.IFraud
	IWalletConfig         repositories.IWalletConfig
}

func NewOrderApplication(config *configs.Config, logger *zap.Logger, pool gpooling.IPool) *OrderApplication {
	opts := rabbitmq.NewOptions().WithUri(config.QueueUri)

	queue, _ := rabbitmq.NewRabbiMQ(*opts, *config, logger, pool)
	db := database_mgo.NewMongoDBconnection(config.MongoURI)

	//Original GRPC
	service := micro.NewService()

	//ETCD
	service.Init()
	servicePromotion := service_promotion.NewPromotionServicesService(config.Services.ServicePromotion, service.Client())
	serviceTransaction := service_transaction.NewService(config.Services.ServiceTransaction, service.Client())
	serviceUser := service_user.NewService(config.Services.ServiceUser, service.Client())
	serviceMerchantFee := service_merchant_fee.NewService(config.Services.ServiceMerchantFee, service.Client())

	var mqttObjClient = []mqttCl.Client{
		mqtt.Connection(config.MQTTMobileUri, "admin", "admin"),
		mqtt.Connection(config.MQTTInternalUri.Uri, config.MQTTInternalUri.Username, config.MQTTInternalUri.Password),
	}

	application := &OrderApplication{
		Config:                config,
		Queue:                 queue,
		Logger:                logger,
		OrderRepository:       order.NewOrderCollectionImpl(db, config),
		TransactionRepository: serviceTransaction,
		PromotionRepository:   servicePromotion,
		UserRepository:        serviceUser,
		MerchantFeeRepository: serviceMerchantFee,
		BankServiceRepository: bank_service.NewRepoImpl(config.BankGwURI, logger),
		IPool:                 pool,
		LogSaga:               saga.New(),
		LixiRepository:        lixi.NewLixiCollectionImpl(db),
		MQTT:                  mqtt.NewMQTTRepositoryImpl(mqttObjClient, logger),
		IMessage:              messages.NewRepository(db),
		IDevice:               devices.NewRepository(db),
		IFirebase:             firebase.NewRepoImpl(config.FirebaseKey),
		IVA:                   va_VCCB.NewVAVCCBRepository(db, "va_vccb", config),
		IFraud:                fraud.NewRepoImpl(config.FraudUri, logger),
		IWalletConfig:         wallet_config.NewServiceWalletConfigRepository(db, "wallet", config),
	}
	conn, _ := grpc.Dial(config.Services.ServiceCard, grpc.WithInsecure())
	application.ServiceCardRepository = service_card.NewServiceCardClient(conn)

	application.RegisterConsumerTopic([]string{constants.TopicUpdateVABalance, constants.TopicUpdateBankStatus})

	return application
}
