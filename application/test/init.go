package test

import (
	"go.mongodb.org/mongo-driver/mongo"
	"orders-system/application"
	"orders-system/domain/repositories/mocks"
	"orders-system/infrastructure/database_mgo"
	"orders-system/infrastructure/database_mgo/order"
	"orders-system/infrastructure/rabbitmq"
	mock_service_card "orders-system/proto/service_card/mocks"
	mock_service_merchant_fee "orders-system/proto/service_merchant_fee/mocks"
	mock_prom "orders-system/proto/service_promotion/mocks"
	mock_trans "orders-system/proto/service_transaction/mocks"
	mock_users "orders-system/proto/service_user/mocks"
	"orders-system/utils/saga"

	"orders-system/utils/configs"
	"orders-system/utils/gpooling"
	logger2 "orders-system/utils/logger"
)

type MockService struct {
	DB                 *mongo.Database
	OrderRepository    *mocks.OrderRepository
	OrderApplication   *application.OrderApplication
	Transaction        *mock_trans.Service
	Promotion          *mock_prom.PromotionServicesService
	Card               *mock_service_card.ServiceCardClient
	User               *mock_users.Service
	Bank               *mocks.BankRepository
	BankService        *mocks.BankServiceRepository
	MerchantFeeService *mock_service_merchant_fee.Service
	Lixi               *mocks.LixiRepository
	Mqtt               *mocks.IMqtt
	IVA                *mocks.IVA
	IWalletConfig      *mocks.IWalletConfig
}

func NewTestOrderApplication() *MockService {
	config, err := configs.LoadTestConfig("../../")

	if err != nil {
		panic(err)
	}

	logger, err := logger2.NewLogger("production")

	if err != nil {
		panic(err)
	}

	pool, err := gpooling.NewPooling(config.MaxPoolSize)

	if err != nil {
		panic(err)
	}

	opts := rabbitmq.NewOptions().WithUri(config.QueueUri)

	queue, err := rabbitmq.NewRabbiMQ(*opts, *config, logger, pool)

	if err != nil {
		panic(err)
	}

	db := database_mgo.NewMongoDBconnection(config.MongoURI)
	transaction_service := &mock_trans.Service{}
	prom_service := &mock_prom.PromotionServicesService{}
	card_service := &mock_service_card.ServiceCardClient{}
	user_service := &mock_users.Service{}
	merchantFeeService := &mock_service_merchant_fee.Service{}
	mqttMock := &mocks.IMqtt{}

	bank := &mocks.BankRepository{}
	bankService := &mocks.BankServiceRepository{}
	orderRepoMock := &mocks.OrderRepository{}
	lixi := &mocks.LixiRepository{}
	iVA := &mocks.IVA{}

	iWalletConfig := &mocks.IWalletConfig{}

	return &MockService{
		DB:              db.Database("service_order_system"),
		OrderRepository: orderRepoMock,
		OrderApplication: &application.OrderApplication{
			Config:                config,
			Queue:                 queue,
			Logger:                logger,
			OrderRepository:       order.NewOrderCollectionImpl(db, config),
			TransactionRepository: transaction_service,
			PromotionRepository:   prom_service,
			UserRepository:        user_service,
			MerchantFeeRepository: merchantFeeService,
			ServiceCardRepository: card_service,
			BankRepository:        bank,
			BankServiceRepository: bankService,
			IPool:                 pool,
			LogSaga:               saga.New(),
			LixiRepository:        lixi,
			MQTT:                  mqttMock,
			IVA:                   iVA,
			IWalletConfig:         iWalletConfig,
		},
		Transaction:        transaction_service,
		Promotion:          prom_service,
		Card:               card_service,
		User:               user_service,
		Bank:               bank,
		BankService:        bankService,
		MerchantFeeService: merchantFeeService,
		Lixi:               lixi,
		Mqtt:               mqttMock,
		IVA:                iVA,
		IWalletConfig:      iWalletConfig,
	}
}
