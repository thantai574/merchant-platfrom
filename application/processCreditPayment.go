package application

import (
	"context"
	"errors"
	"github.com/golang/protobuf/ptypes"
	"go.uber.org/zap"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	eBankGw "orders-system/domain/entities/bank_gateway"
	"orders-system/domain/value_objects"
	mapErr "orders-system/errors"
	"orders-system/proto/order_system"
	"orders-system/proto/service_transaction"
	"orders-system/utils/saga"
	"orders-system/utils/telegram"
	"strconv"
)

func (us *OrderApplication) ProcessCreditPayment(ctx context.Context, request *order_system.ProcessCreditPaymentRequest, response *order_system.ProcessCreditPaymentResponse) (orderDto *entities.OrderEntity, err error) {
	sg := saga.NewSaga("ProcessCreditPayment Order")

	var sof string

	checkBankBin, err := us.BankServiceRepository.CheckInternationalBankBin(request.CardNumber[0:6])
	if err != nil || len(checkBankBin.Data.Data) == 0 {
		return orderDto, errors.New("Hệ thống chưa hỗ trợ thanh toán thẻ trên")
	}

	if checkBankBin.Data.Data[0].CardType.IsCredit() {
		sof = constants.SOURCE_OF_FUND_CREDIT_CARD
	}
	if checkBankBin.Data.Data[0].CardType.IsDebit() {
		sof = constants.SOURCE_OF_FUND_DEBIT_CARD
	}

	var trans service_transaction.ETransactionDTO
	getOrderById, err := us.GetValidOrder(ctx, request.OrderId)
	if err != nil {
		return orderDto, err
	}

	getOrderById.SourceOfFund = sof
	orderDto = getOrderById

	//todo check fraud
	checkFraudRes, err := us.IFraud.GetFraud(request.CardNumber)
	if err != nil {
		us.Logger.With(zap.Error(err)).Error("err_get_fraud")
		checkFraudRes = entities.Fraud{}
	}

	if orderDto.TransactionID == "" {
		initTrans := &service_transaction.ETransactionDTO{
			ServiceType:          orderDto.ServiceType,
			TransactionType:      orderDto.OrderType,
			Amount:               orderDto.Amount,
			SourceOfFund:         orderDto.SourceOfFund,
			RefId:                orderDto.RefID,
			MerchantID:           orderDto.SubscribeMerchantID,
			SubscriberMerchantID: orderDto.SubscribeMerchantID,
			BankCode:             constants.VPB,
			OrderId:              orderDto.OrderID,
			MerchantTypeWallet:   constants.AmountRevenue,
			IbftReceiveBank:      checkBankBin.Data.Data[0].BankName,
			AppId:                orderDto.ServiceCode,
		}
		initTrans, err = us.serviceTransactionInit(ctx, initTrans)
		if err != nil {
			return orderDto, err
		}
		trans = *initTrans
	} else {
		getTransByID, err := us.TransactionRepository.FindTransactionByID(context.TODO(), &service_transaction.ETransactionDTO{
			TransactionId: orderDto.TransactionID,
		})

		if err != nil {
			us.Logger.With(zap.Error(err)).Error("err_get_tran_mpgs")
			return orderDto, err
		}
		trans = *getTransByID
	}

	//todo process Order
	err = sg.AddStep(&saga.Step{
		Name: "PROCESSING_ORDER",
		Func: func(c context.Context) (err error) {
			if orderDto.TransactionID == "" {
				orderDto.TransactionID = trans.TransactionId
			}
			orderDto, err = us.ProcessingOrder(ctx, orderDto)
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			return
		},
		Options: nil,
	})
	if err != nil {
		return
	}

	//todo Fail case
	err = sg.AddStep(&saga.Step{
		Name: "FAIL_CASE_FRAUD",
		Func: func(c context.Context) (err error) {
			if checkFraudRes.Status.IsFail() {
				trans.FailReason = checkFraudRes.Reason
				trans.IbftReceiveBank = checkBankBin.Data.Data[0].BankName
				_, err = us.serviceTransactionCancel(ctx, &trans)
				if err != nil {
					us.Logger.With(zap.Error(err)).Error("err_cancel_mpgs")
					return err
				}

				orderDto.FailReason = trans.FailReason
				_, err = us.FailedOrder(ctx, orderDto)

				err = us.CreateMessageMqtt(context.TODO(), constants.TopicMQTTFailFraudMPGS, constants.MQTTEventBackground, constants.TopicMQTTFailFraudMPGS, trans, false)
				err = us.warningBlackList(*orderDto, trans, checkFraudRes)
				return mapErr.ErrFailOrder
			}
			return err
		},
		CompensateFunc: func(c context.Context) (err error) {
			return
		},
		Options: nil,
	})

	//todo call bank
	err = sg.AddStep(&saga.Step{
		Name: "PROCESS CREDIT PAYMENT ORDER",
		Func: func(c context.Context) (err error) {
			amount := strconv.FormatInt(getOrderById.Amount, 10)
			callBackCreditPaymentResponse, err := us.BankServiceRepository.CreditPayment(eBankGw.CreditPaymentRequestData{
				CardNumber:        request.CardNumber,
				GpayTransactionId: request.OrderId,
				Amount:            amount,
				ExpiryYear:        request.ExpiryYear,
				ExpiryMonth:       request.ExpiryMonth,
				SecurityCode:      request.SecurityCode,
				CardHolderName:    request.CardHolderName,
				RedirectUrl:       request.RedirectUrl,
				GpayUserId:        orderDto.UserID,
				Token:             request.Token,
				MerchantCode:      orderDto.MerchantCode,
				MCC:               orderDto.MerchantCategoryCode,
				MccType:           orderDto.MerchantCategoryType,
			})
			if err != nil {
				return err
			}

			response.OrderId = request.OrderId
			response.AcsUrl = callBackCreditPaymentResponse.Data.TripleDSecure.AuthenticationRedirect.Customized.AcsUrl
			response.PaReq = callBackCreditPaymentResponse.Data.TripleDSecure.AuthenticationRedirect.Customized.PaReq
			response.Md = callBackCreditPaymentResponse.Data.Md
			response.TermUrl = callBackCreditPaymentResponse.Data.CallbackUrl

			expiredAt, err := ptypes.TimestampProto(orderDto.ExpiredAt)
			if err == nil {
				response.ExpiredAt = expiredAt
			}

			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			return
		},
		Options: nil,
	})
	if err != nil {
		return
	}

	//todo send telegram
	err = sg.AddStep(&saga.Step{
		Name: "PROCESS CREDIT PAYMENT ORDER",
		Func: func(c context.Context) (err error) {
			_ = us.warningBlackList(*orderDto, trans, checkFraudRes)
			return err
		},
		CompensateFunc: func(c context.Context) (err error) {
			return
		},
		Options: nil,
	})

	ordinator := saga.NewCoordinator(ctx, ctx, sg, us.LogSaga)
	rg := ordinator.Play()
	err = rg.ExecutionError
	return
}

func (us *OrderApplication) warningBlackList(orderDto entities.OrderEntity, trans service_transaction.ETransactionDTO, checkFraudRes entities.Fraud) (err error) {
	if checkFraudRes.Status.IsSuccess() || checkFraudRes.Status.IsFail() {
		defer us.IPool.Submit(func() {
			orderInfoSend := telegram.SendOrderFraud(orderDto, trans, checkFraudRes)
			telegram.SendTelegram(orderInfoSend, us.Config.TelegramChannelId.Fraud)

		})
	}

	_, err = us.IFraud.SaveFraud(value_objects.FraudTransRequest{
		TransactionId: trans.Id,
		FraudDto:      checkFraudRes,
	})
	if err != nil {
		us.Logger.Error("err_save_fraud_service", zap.Error(err))
	}
	return err
}
