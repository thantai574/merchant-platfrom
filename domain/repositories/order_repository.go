package repositories

import (
	"context"
	"orders-system/domain/entities"
	eBankGw "orders-system/domain/entities/bank_gateway"
	"orders-system/domain/request_params"
	"orders-system/domain/value_objects"
	"orders-system/proto/order_system"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.mongodb.org/mongo-driver/bson"
)

type OrderRepository interface {
	Create(ctx context.Context, entity *entities.OrderEntity) (*entities.OrderEntity, error)
	FindByOrderID(ctx context.Context, orderID string) (res *entities.OrderEntity, err error)
	FindByMerchantIdAndRefId(ctx context.Context, merchantId, refId string) (res *entities.OrderEntity, err error)
	FindByID(ctx context.Context, orderID string) (*entities.OrderEntity, error)
	ReplaceByID(ctx context.Context, entity *entities.OrderEntity) (*entities.OrderEntity, error)
	ProcessingOrderByID(ctx context.Context, entity *entities.OrderEntity) (*entities.OrderEntity, error)
	CheckLuckyMoney(ctx context.Context, user_id, lucky_money_id string) (*entities.OrderEntity, error)
	GetExpiredOrder(ctx context.Context) ([]*entities.OrderEntity, error)
	GetOrderByRefundId(ctx context.Context, refundId string) (*entities.OrderEntity, error)
	GetOrderMerchant(ctx context.Context, req order_system.GetOrderByMerchantReq) (*order_system.GetOrderByMerchantRes, error)
}

type BankRepository interface {
	GetUserLinkedList(userID string) (linkedList []*entities.LinkedBankLink, err error)
	GetDetailLink(id string) (linkedList *entities.LinkedBankLink, err error)
}

type BankServiceRepository interface {
	CheckGpayBankBalance(gpayBankCode, accountNumber string) (eBankGw.BalanceCheckResponse, error)
	IBFTCheckBalance(accountNumber string) (eBankGw.IBFTGpayBalanceCheckResponse, error)

	LinkInfo(linkId int64) (eBankGw.LinkInfoResponse, error)
	LinkList(gpayUserId string) (eBankGw.ListLinkResponse, error)

	Link(data eBankGw.LinkRequestData, gpayBankCode, clientIP string) (eBankGw.LinkResponse, error)
	UnLink(bankCode string, linkId int64) (eBankGw.UnLinkResponse, error)
	VerifyOTP(gpayBankCode, refBankTraceID, order_id string, linkId int64, otp string) (eBankGw.VerifyOTPResponse, error)
	CashIn(gpayBankCode string, linkId, amount int64, gpayOrderId, userId string) (eBankGw.CashInResponse, error)              // nạp tiền vào ví GPAY
	CashInNapas(request eBankGw.NapasCashInDataRequest, gpayBankCode, clientIP string) (eBankGw.NapasCashInResponse, error)    // nạp tiền NAPAS
	CashOut(amount int64, bankCode, orderId string, linkId int64, description, userId string) (eBankGw.CashOutResponse, error) // rút tiền khỏi ví GPAY
	RetrieveOrderStatus(orderId string, bankCode string) (eBankGw.RetrieveCreditOrderResponse, error)
	ReFund(orderId string, amount int64) (eBankGw.RefundRes, error)
	CheckByPassOTP(req eBankGw.CheckByPassOTPDataReq) (eBankGw.CheckByPassOTPDataRes, error)

	CreditPayment(eBankGw.CreditPaymentRequestData) (eBankGw.CreditPaymentResponse, error) // credit payment
	CheckInternationalBankBin(bankbin string) (eBankGw.CheckBankBinResponse, error)

	IBFTInquiry(accountNumber, cardNumber, ibftCode string) (eBankGw.IBFTInquiryCheckResponse, error)                                                               // check thông tin người nhận IBFT
	IBFTTransfer(accountNumber, cardNumber, gpayUserId, orderId, ibftCode, description string, amount int64, bankName string) (eBankGw.IBFTTransferResponse, error) // CK IBFT

	CreateVA(requestData eBankGw.CreateVARequestData, provider string) (eBankGw.ActionVAResponse, error)
	ReOpenVA(req eBankGw.ReOpenVARequestData, provider string) (eBankGw.ActionVAResponse, error)
	UpdateVA(requestData eBankGw.CreateVARequestData, provider string) (eBankGw.ActionVAResponse, error)
	CloseVA(req eBankGw.CloseVARequestData, provider string) (eBankGw.ActionVAResponse, error)
	DetailVA(accountNumber string, provider string) (eBankGw.ActionVAResponse, error)

	PGInitTrans(bankCode string, data eBankGw.PGInitTransReqData) (eBankGw.PGInitTransRes, error)
	PGInitOrder(req eBankGw.PGInitTransReq) (eBankGw.PGInitOrderRes, error)

	GetBanks() (eBankGw.BankRes, error)
}

type LixiRepository interface {
	FindStarted(ctx context.Context) (res []*entities.LixiEntity, err error)
	UpdateOne(ctx context.Context, lixi *entities.LixiEntity) (lixi_response *entities.LixiEntity, err error)
	FindRandomAmount(ctx context.Context, min, max int64) (res *entities.LixiAmount, err error)
}

type IMqtt interface {
	Publish(topic, message string, retain bool, prefix string) error
	Subscribe(topic string, c func(client mqtt.Client, message mqtt.Message))
}

type IDevice interface {
	CreateDevice(devices entities.Devices) (*entities.Devices, error)

	FindById(id string) (*entities.Devices, error)

	FindByUserId(user_id string) (*entities.Devices, error)

	DeleteDevice(device_id, user_id string) error

	UpdateDevice(device_id, device_token string) error
}

type IMessage interface {
	FindById(id string) (*entities.Message, error)

	ListMessage(offset int64, topics []string) ([]*entities.Message, error)

	CountMessageStatusUnread(uid string) (int64, error)

	CreateMessage(entities.Message) (*entities.Message, error)
}

type IVA interface {
	CreateVA(request entities.VirtualAccounts) (entities.VirtualAccounts, error)
	GetVAByMerchantId(merchantId string) (entities.VirtualAccounts, error)
	GetVAByMapCondition(mapId, mapType, merchantId string) (entities.VirtualAccounts, error)
	GetVAByAccountNumber(accountNumber string) (entities.VirtualAccounts, error)
	UpdateVA(accountNumber string, fieldUpdate bson.M) (entities.VirtualAccounts, error)
	IncrementBalanceVA(accountNumber string, balance int64) (entities.VirtualAccounts, error)
	DeleteVAAccount(accountNumber string) error
}

type IWalletConfig interface {
	GetMerchantConfig(ctx context.Context, req request_params.GetMerchantConfigReq) (value_objects.GetMerchantConfigRes, error)
	GetRefundConfig(ctx context.Context, req request_params.GetRefundConfigReq) (value_objects.GetRefundConfigRes, error)
}

type IFraud interface {
	GetFraud(cardNumber string) (entities.Fraud, error)
	SaveFraud(request value_objects.FraudTransRequest) (entities.Fraud, error)
}
