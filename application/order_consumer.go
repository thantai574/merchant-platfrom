package application

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	"orders-system/errors"
	errorsMap "orders-system/errors"
	order_system "orders-system/proto/order_system"
	"orders-system/proto/service_card"
	"orders-system/proto/service_merchant_fee"
	"orders-system/proto/service_promotion"
	"orders-system/proto/service_transaction"
	mapErrString "orders-system/utils/errors"
	"orders-system/utils/helpers"
	"orders-system/utils/saga"
	"strconv"
	"strings"
	"time"
)

func (us *OrderApplication) ConsumerCreateOrder(msg []byte) error {
	msgData := &order_system.Request{}

	err := proto.Unmarshal(msg, msgData)
	if err == nil {
		us.Logger.Info(msgData.Name)
	}

	return err
}

func (us *OrderApplication) RegisterConsumerTopic(topics []string) {
	for _, topic := range topics {
		switch topic {
		case constants.TopicUpdateBankStatus:
			_ = us.Queue.WithConsumerTopic(us.UpdateBankStatus, topic)
		case constants.TopicUpdateVABalance:
			_ = us.Queue.WithConsumerTopic(us.UpdateBalanceVA, topic)
		}

	}

}

func (us *OrderApplication) UpdateBalanceVA(ctx context.Context, msg []byte) (orderDto *entities.OrderEntity, err error) {
	msgData := &order_system.VAChangeBalanceResponse{}

	err = proto.Unmarshal(msg, msgData)

	if err == nil {
		us.Logger.With(zap.Reflect("value", msgData)).Info("msg_data")

		amount, _ := strconv.ParseInt(msgData.Amount, 10, 64)
		incrementBalanceVARes, err := us.IVA.IncrementBalanceVA(msgData.AccountNumber, amount)
		if err != nil {
			us.Logger.With(zap.Reflect("err", err)).Error("update_va_balance_fail")
			return orderDto, err
		}

		if incrementBalanceVARes.AccountType == constants.VA_ACCOUNT_TYPE_ONCE {
			_, err = us.IVA.UpdateVA(msgData.AccountNumber, bson.M{"status": "CLOSE"})
			if err == nil {
				_ = us.CreateMessageMqtt(context.TODO(), constants.TopicMQTTCloseVA, constants.MQTTEventBackground, constants.TopicMQTTCloseVA, incrementBalanceVARes, false)
			}
		}

		sg := saga.NewSaga("Accounting VA Balance Change Order")
		orderDto := new(entities.OrderEntity)

		err = sg.AddStep(&saga.Step{
			Name: "INIT_ORDER",
			Func: func(c context.Context) (err error) {
				orderDto, err = us.InitOrder(ctx, &entities.OrderEntity{
					RefID:               msgData.BankTraceId,
					SubscribeMerchantID: incrementBalanceVARes.MerchantId,
					ServiceType:         constants.SERVICE_TYPE_COLLECTION_AND_PAY,
					OrderType:           constants.TRANSTYPE_PAY_VA,
					Amount:              amount,
					MerchantCode:        incrementBalanceVARes.MerchantCode,
					AccountNo:           msgData.AccountNumber,
					BankTransactionId:   msgData.BankTransactionId,
					Description:         msgData.Description,
					BankCode:            incrementBalanceVARes.Provider,
					AccountName:         incrementBalanceVARes.AccountName,
				})

				return
			},
			CompensateFunc: func(c context.Context) (err error) {
				if !orderDto.Status.IsVerifying() && !orderDto.Status.IsFailed() {
					us.FailedOrder(ctx, orderDto)
				}
				return
			},
			Options: nil,
		})
		if err != nil {
			return orderDto, err
		}

		err = sg.AddStep(&saga.Step{
			Name: "PROCESSING_ORDER",
			Func: func(c context.Context) (err error) {
				orderDto, err = us.ProcessingOrder(ctx, orderDto)
				return
			},
			CompensateFunc: func(c context.Context) (err error) {
				if !orderDto.Status.IsVerifying() && !orderDto.Status.IsFailed() {
					us.FailedOrder(ctx, orderDto)
				}
				return
			},
			Options: nil,
		})
		if err != nil {
			return orderDto, err
		}

		var trans *service_transaction.ETransactionDTO
		var failReason string

		//@todo Trans Init
		err = sg.AddStep(&saga.Step{
			Name: "INIT_PAYMENT",
			Func: func(c context.Context) (err error) {
				trans, err = us.serviceTransactionInit(ctx, &service_transaction.ETransactionDTO{
					Message:            orderDto.Description,
					Status:             incrementBalanceVARes.MapType,
					ServiceType:        constants.SERVICE_TYPE_COLLECTION_AND_PAY,
					TransactionType:    constants.TRANSTYPE_PAY_VA,
					Amount:             amount,
					MerchantID:         incrementBalanceVARes.MerchantId,
					RefId:              msgData.BankTraceId,
					OrderId:            orderDto.OrderID,
					BankTransactionId:  msgData.BankTransactionId,
					BankCode:           incrementBalanceVARes.Provider,
					AccountVA:          msgData.AccountNumber,
					BankCodeVA:         incrementBalanceVARes.Provider,
					CustomerNameVA:     incrementBalanceVARes.AccountName,
					TimeSendVA:         msgData.Time / 1000,
					MerchantTypeWallet: constants.AmountRevenue,
					IdentityVANumber:   incrementBalanceVARes.MapId,
					AccountVAType:      incrementBalanceVARes.AccountType,
				})

				if err != nil {
					us.Logger.With(zap.Error(err)).Error("SERIVCE_TRANSACTION.err_init")
					failReason = err.Error()
					return err
				}
				return
			},
			CompensateFunc: func(c context.Context) (err error) {
				if trans != nil && trans.Status != constants.TRANSACTION_STATUS_FAILED && trans.Status != constants.TRANSACTION_STATUS_PENDING {
					trans.FailReason = failReason
					trans, err = us.serviceTransactionCancel(ctx, trans)
				}
				if !orderDto.Status.IsVerifying() && !orderDto.Status.IsFailed() {
					orderDto.InternalErr = failReason
					_, _ = us.FailedOrder(c, orderDto)
				}
				return
			},
		})
		if err != nil {
			return orderDto, err
		}

		err = sg.AddStep(&saga.Step{
			Name: "FINALIZATION_PAYMENT",
			Func: func(c context.Context) (err error) {
				trans, err = us.serviceTransactionConfirm(ctx, trans)
				return
			},
			CompensateFunc: func(c context.Context) (err error) {
				if trans.Status != constants.TRANSACTION_STATUS_PENDING {
					trans, err = us.serviceTransactionPending(ctx, trans)
				}
				if !orderDto.Status.IsVerifying() {
					_, _ = us.VerifyingOrder(ctx, orderDto)
				}
				return
			},
		})
		if err != nil {
			return orderDto, err
		}

		err = sg.AddStep(&saga.Step{
			Name: "SUCCESS_ORDER",
			Func: func(c context.Context) (err error) {
				orderDto.TransactionID = trans.TransactionId
				orderDto.SucceedAt = time.Unix(msgData.Time/1000, 0)
				orderDto, err = us.SuccessOrder(ctx, orderDto)
				if err == nil {
					us.IPool.Submit(func() {
						_ = us.CreateMessageMqtt(context.TODO(), constants.TopicMQTTUpdateBalanceVA, constants.MQTTEventBackground, constants.TopicMQTTUpdateBalanceVA, trans, false)
					})
				}
				return
			},
			CompensateFunc: func(c context.Context) (err error) {
				if trans.Status != constants.TRANSACTION_STATUS_PENDING {
					trans, err = us.serviceTransactionPending(c, trans)
				}
				if !orderDto.Status.IsVerifying() {
					us.VerifyingOrder(ctx, orderDto)
				}
				return
			},
			Options: nil,
		})
		if err != nil {
			return orderDto, err
		}

		ordinator := saga.NewCoordinator(ctx, ctx, sg, us.LogSaga)
		rg := ordinator.Play()
		err = rg.ExecutionError
		return orderDto, err

	} else {
		us.Logger.With(zap.Reflect("value", msg)).Error("va_queue_balance_err", zap.Error(err))
	}

	return
}

func (us *OrderApplication) UpdateBankStatus(ctx context.Context, msg []byte) (orderDto *entities.OrderEntity, err error) {
	msgData := &order_system.BankOrderStatusResponse{}

	err = proto.Unmarshal(msg, msgData)

	if err == nil {
		us.Logger.With(zap.Reflect("value", msgData)).Info("msg_data")

		sg := saga.NewSaga("UpdateBankStatus")
		orderDto := new(entities.OrderEntity)
		bankTraceId := msgData.BankTransactionId

		err = sg.AddStep(&saga.Step{
			Name: "Get_ORDER",
			Func: func(c context.Context) (err error) {
				findOrder, err := us.GetValidOrder(ctx, msgData.OrderId)
				if err != nil {
					return err
				}
				orderDto = findOrder
				return
			},
			CompensateFunc: func(c context.Context) (err error) {
				return
			},
			Options: nil,
		})
		if err != nil {
			return orderDto, err
		}

		var trans *service_transaction.ETransactionDTO
		var failReason, voucherId string

		//@todo Trans Init
		err = sg.AddStep(&saga.Step{
			Name: "FIND_TRANS",
			Func: func(c context.Context) (err error) {
				findTranById, err := us.TransactionRepository.FindTransactionByID(ctx, &service_transaction.ETransactionDTO{TransactionId: orderDto.TransactionID})
				if err != nil {
					return err
				}

				findTranById.BankTransactionId = orderDto.BankTransactionId
				trans = findTranById
				return
			},
			CompensateFunc: func(c context.Context) (err error) {
				if trans != nil && trans.Status != constants.TRANSACTION_STATUS_FAILED && trans.Status != constants.TRANSACTION_STATUS_PENDING {
					trans.FailReason = failReason
					trans.BankTransactionId = bankTraceId
					trans, err = us.serviceTransactionCancel(ctx, trans)
				}
				if !orderDto.Status.IsVerifying() && !orderDto.Status.IsFailed() {
					orderDto.InternalErr = failReason
					orderDto.BankTransactionId = bankTraceId
					_, _ = us.FailedOrder(ctx, orderDto)
				}
				return err
			},
		})
		if err != nil {
			return orderDto, err
		}

		//@todo FAIL BANK STATUS
		err = sg.AddStep(&saga.Step{
			Name: "FAIL BANK STATUS",
			Func: func(c context.Context) (err error) {
				if msgData.ErrorCode != "200" {
					failReason = msgData.Description
					trans.FailReason = failReason
					trans.RefId = msgData.RefId
					trans.BankTransactionId = bankTraceId
					_, err = us.serviceTransactionCancel(ctx, trans)
					if err != nil {
						us.Logger.With(zap.Any("request", trans)).Error(err.Error())
						return err
					}
					orderDto.FailReason = failReason
					_, _ = us.SuccessOrder(ctx, orderDto)
					return errorsMap.ErrGeneral
				}

				return err
			},
			CompensateFunc: func(c context.Context) (err error) {
				return errorsMap.ErrGeneral
			},
		})
		if err != nil {
			return orderDto, err
		}

		// todo VOUCHER
		err = sg.AddStep(&saga.Step{
			Name: "ACCOUNT_VOUCHER",
			Func: func(c context.Context) (err error) {
				if orderDto.VoucherCode != "" {
					res, err := us.servicePromotionUsed(ctx, &service_promotion.UseVoucherRequest{
						Code:         orderDto.VoucherCode,
						UserId:       orderDto.UserID,
						TraceId:      orderDto.OrderID,
						Total:        1,
						Amount:       orderDto.Amount,
						ServiceCode:  orderDto.SubOrderType,
						SourceOfFund: orderDto.SourceOfFund,
					})
					if err != nil {
						return err
					}
					voucherId = res.Voucher.Voucher.Id
					trans.VoucherID = voucherId
				}
				return err
			},
			CompensateFunc: func(c context.Context) (err error) {
				_, err = us.servicePromotionCompensate(ctx, &service_promotion.ReverseWalletRequest{
					TraceId: orderDto.OrderID,
				})
				return
			},
			Options: nil,
		})
		if err != nil {
			return orderDto, err
		}

		isPendingSrvCard := false
		var providerMerchantId, providerMerchant, cardType string

		// todo BuyCard (BUYCARD+ BUYDATA+ BUYCARDGAME)
		err = sg.AddStep(&saga.Step{
			Name: "BuyCard",
			Func: func(c context.Context) (err error) {
				if helpers.IsStringSliceContains([]string{constants.SUB_TRANSTYPE_WALLET_BUY_DATA,
					constants.SUB_TRANSTYPE_WALLET_BUY_CARD,
				}, orderDto.SubOrderType) {
					buycardRes, err := us.ServiceCardRepository.BuyCard(ctx, &service_card.BuyCardReq{
						Price:         orderDto.Amount,
						Telco:         orderDto.OrderCardTelco,
						Quantity:      orderDto.Quantity,
						UserId:        orderDto.UserID,
						OrderId:       orderDto.OrderID,
						TransactionId: trans.TransactionId,
					})
					if err != nil {
						failReason = mapErrString.GetGrpcErrMessage(err)
						return fmt.Errorf("%v", failReason)
					}

					us.Logger.Named(fmt.Sprintf("%v", buycardRes)).Info("buyCardResponse")

					providerMerchantId = buycardRes.ProviderMerchantId
					providerMerchant = buycardRes.ProviderMerchant
					cardType = buycardRes.CardType

					trans.ProviderMerchantID = providerMerchantId

					if buycardRes != nil && buycardRes.Status == constants.TRANSACTION_STATUS_PENDING {
						isPendingSrvCard = true
						return errors.ErrPendingOrder
					}
				}
				return
			},
			CompensateFunc: func(c context.Context) (err error) {
				if isPendingSrvCard == true { // Pending Order Service Card
					if trans.Status != constants.TRANSACTION_STATUS_PENDING {
						trans, err = us.serviceTransactionPending(ctx, trans)
					}
					if !orderDto.Status.IsVerifying() {
						_, _ = us.VerifyingOrder(ctx, orderDto)
					}
				} else {
					trans.FailReason = failReason
					trans, err = us.serviceTransactionCancel(ctx, trans)

					orderDto.InternalErr = failReason
					_, err = us.FailedOrder(ctx, orderDto)
				}
				return err
			},
		})
		if err != nil {
			return orderDto, err
		}

		//todo Accounting Merchant Fee (BUYCARD+ BUYDATA+ BUYCARDGAME)
		err = sg.AddStep(&saga.Step{
			Name: "PROVIDER CARD FEE",
			Func: func(c context.Context) (err error) {
				if helpers.IsStringSliceContains([]string{constants.SUB_TRANSTYPE_WALLET_BUY_DATA,
					constants.SUB_TRANSTYPE_WALLET_BUY_CARD,
				}, orderDto.SubOrderType) {
					var cardNameMerchantQuotaCheck string
					if strings.Contains(orderDto.OrderCardTelco, "DT_") {
						cardNameMerchantQuotaCheck = strings.Replace(orderDto.OrderCardTelco, "DT_", "", -1)
					} else {
						cardNameMerchantQuotaCheck = orderDto.OrderCardTelco
					}

					merchantFee, err := us.MerchantFeeRepository.GetMerchantVendorDiscount(ctx, &service_merchant_fee.GetMerchantVendorDiscountReq{
						Amount:       orderDto.Amount,
						ServiceType:  orderDto.ServiceType,
						TransType:    orderDto.OrderType,
						SubTransType: orderDto.SubOrderType,
						MerchantId:   orderDto.SubscribeMerchantID,
						Card: &service_merchant_fee.GetMerchantVendorDiscountCard{
							Provider: providerMerchant,
							CardName: cardNameMerchantQuotaCheck,
							CardType: cardType,
						},
					})
					if err != nil {
						failReason = err.Error()
						us.Logger.Error("merchant_provider_quota_fee_err", zap.Error(err))
					} else {
						trans.DiscountForMerchantSubscriber = orderDto.Quantity * merchantFee.MerchantDiscountAmount // chiet khau gpay cho merchant
						trans.DiscountOfMerchantProvider = orderDto.Quantity * merchantFee.VendorDiscountAmount      // chiet khau provider ZOTA/IMEDIA cho GPAY
					}
				}
				return
			},
			CompensateFunc: func(c context.Context) (err error) {
				return
			},
		})
		if err != nil {
			return nil, err
		}

		// todo case TOPUP
		err = sg.AddStep(&saga.Step{
			Name: "TOP_UP",
			Func: func(c context.Context) (err error) {
				if orderDto.SubOrderType == constants.SUB_TRANSTYPE_WALLET_TOPUP_CARD {
					topUpRes, err := us.ServiceCardRepository.Topup(ctx, &service_card.TopupReq{
						Price:         orderDto.Amount,
						Telco:         orderDto.OrderCardTelco,
						Phone:         orderDto.PhoneTopUp,
						UserId:        orderDto.UserID,
						OrderId:       orderDto.OrderID,
						SubTransType:  orderDto.SubOrderType,
						TransactionId: trans.TransactionId,
					})

					if err != nil {
						failReason = err.Error()
						return errorsMap.ErrFailOrder
					}

					providerMerchantId = topUpRes.ProviderMerchantId
					trans.ProviderMerchantID = providerMerchantId

					if topUpRes != nil && topUpRes.Status == constants.TRANSACTION_STATUS_PENDING {
						isPendingSrvCard = true
						return errorsMap.ErrPendingOrder
					}
				}

				return err
			},
			CompensateFunc: func(c context.Context) (err error) {
				if isPendingSrvCard == true { // Pending Order Service Card
					if trans.Status != constants.TRANSACTION_STATUS_PENDING {
						trans, err = us.serviceTransactionPending(ctx, trans)
					}
					if !orderDto.Status.IsVerifying() {
						us.VerifyingOrder(ctx, orderDto)
					}
				} else {
					trans.FailReason = failReason
					trans, err = us.serviceTransactionCancel(ctx, trans)

					orderDto.InternalErr = failReason
					_, err = us.FailedOrder(ctx, orderDto)
				}

				return
			},
		})
		if err != nil {
			return nil, err
		}

		// todo case BILL
		err = sg.AddStep(&saga.Step{
			Name: "PAID_BILL",
			Func: func(c context.Context) (err error) {
				if helpers.IsStringSliceContains([]string{constants.SUB_TRANSTYPE_WALLET_PAY_BILL_WATTER,
					constants.SUB_TRANSTYPE_WALLET_PAY_BILL_ELECTRIC,
					constants.SUB_TRANSTYPE_WALLET_PAY_BILL_LOAN,
					constants.SUB_TRANSTYPE_WALLET_PAY_BILL_INTERNET,
					constants.SUB_TRANSTYPE_WALLET_PAY_BILL_TELEPHONE,
					constants.SUB_TRANSTYPE_WALLET_PAY_BILL_TV,
				}, orderDto.SubOrderType) {
					payBillResponse, err := us.ServiceCardRepository.PaidBill(ctx, &service_card.PaidBillReq{
						Amount:            orderDto.Amount,
						ServiceCode:       orderDto.OrderBillServiceCode,
						CustomerReference: orderDto.OrderBillCustomerRef,
						OrderId:           orderDto.OrderID,
						UserId:            orderDto.UserID,
						SubTransType:      orderDto.SubOrderType,
						AreaCode:          orderDto.OrderBillAreaCode,
						TransactionId:     trans.TransactionId,
					})

					if err != nil {
						failReason = err.Error()
						return errorsMap.ErrFailOrder
					}

					providerMerchantId = payBillResponse.ProviderMerchantId
					trans.ProviderMerchantID = providerMerchantId

					if payBillResponse != nil && payBillResponse.Status == constants.TRANSACTION_STATUS_PENDING {
						isPendingSrvCard = true
						return errorsMap.ErrPendingOrder
					}
				}
				return err
			},
			CompensateFunc: func(c context.Context) (err error) {
				if isPendingSrvCard == true { // Pending Order Service Card
					if trans.Status != constants.TRANSACTION_STATUS_PENDING {
						trans, err = us.serviceTransactionPending(ctx, trans)
					}
					if !orderDto.Status.IsVerifying() {
						us.VerifyingOrder(ctx, orderDto)
					}
				} else {
					trans.FailReason = failReason
					trans, err = us.serviceTransactionCancel(ctx, trans)

					orderDto.InternalErr = failReason
					_, err = us.FailedOrder(ctx, orderDto)
				}

				return
			},
		})
		if err != nil {
			return nil, err
		}

		// todo TRANS SUCCESS
		err = sg.AddStep(&saga.Step{
			Name: "FINALIZATION_PAYMENT",
			Func: func(c context.Context) (err error) {
				trans.BankTransactionId = bankTraceId
				trans, err = us.serviceTransactionConfirm(ctx, trans)
				return
			},
			CompensateFunc: func(c context.Context) (err error) {
				if trans.Status != constants.TRANSACTION_STATUS_PENDING {
					trans, err = us.serviceTransactionPending(ctx, trans)
				}
				if !orderDto.Status.IsVerifying() {
					_, _ = us.VerifyingOrder(ctx, orderDto)
				}
				return
			},
		})
		if err != nil {
			return orderDto, err
		}

		// todo ORDER SUCCESS
		err = sg.AddStep(&saga.Step{
			Name: "SUCCESS_ORDER",
			Func: func(c context.Context) (err error) {
				orderDto, err = us.SuccessOrder(ctx, orderDto)
				return
			},
			CompensateFunc: func(c context.Context) (err error) {
				if trans.Status != constants.TRANSACTION_STATUS_PENDING {
					trans, err = us.serviceTransactionPending(c, trans)
				}
				if !orderDto.Status.IsVerifying() {
					us.VerifyingOrder(ctx, orderDto)
				}
				return
			},
			Options: nil,
		})
		if err != nil {
			return orderDto, err
		}

		ordinator := saga.NewCoordinator(ctx, ctx, sg, us.LogSaga)
		rg := ordinator.Play()
		err = rg.ExecutionError
		return orderDto, err

	} else {
		us.Logger.With(zap.Reflect("value", msg)).Error("get_bank_queue_err", zap.Error(err))
	}

	return
}
