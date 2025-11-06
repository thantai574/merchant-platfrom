package application

import (
	"context"
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	eBank "orders-system/domain/entities/bank_gateway"
	"orders-system/proto/order_system"
	"orders-system/proto/service_transaction"
	"orders-system/utils/saga"
)

func (us *OrderApplication) PayInternationalCard(ctx context.Context, request *order_system.PayInternationalCardRequest) (response *order_system.PayInternationalCardResponse, err error) {
	sg := saga.NewSaga("PayInternational Card")

	var g errgroup.Group
	response = new(order_system.PayInternationalCardResponse)
	var amount = request.Amount

	var discountAmount, userFee int64
	var orderDto *entities.OrderEntity
	var failReason, bankTraceId string
	var trans *service_transaction.ETransactionDTO
	var ibftReceiveBank string

	if request.Quantity > 1 {
		amount = request.Amount * request.Quantity
	}

	var sof = constants.SOURCE_OF_FUND_CREDIT_CARD

	getLinkInfo, err := us.BankServiceRepository.LinkInfo(request.LinkId)
	if err != nil {
		return nil, err
	}

	if getLinkInfo.Data.FundingMethod == constants.MPGS_DEBIT_CARD {
		sof = constants.SOURCE_OF_FUND_DEBIT_CARD
	}

	if request.GetTransType() == constants.TRANSTYPE_WALLET_CASH_IN && sof == constants.SOURCE_OF_FUND_CREDIT_CARD {
		return nil, errors.New("Thẻ không được hỗ trợ nạp tiền")
	}

	g.Go(func() error {
		checkBankBin, err := us.BankServiceRepository.CheckInternationalBankBin(getLinkInfo.Data.CardNumber[0:6])
		if err != nil || len(checkBankBin.Data.Data) == 0 {
			return errors.New("Hệ thống chưa hỗ trợ thanh toán thẻ trên")
		}
		ibftReceiveBank = checkBankBin.Data.Data[0].BankName
		return err
	})

	g.Go(func() error {
		checkFraudRes, err := us.IFraud.GetFraud(getLinkInfo.Data.CardNumber)
		if err != nil || checkFraudRes.Status.IsFail() {
			us.Logger.With(zap.Error(err)).Error("err_get_fraud")
			return errors.New("Hệ thống chưa hỗ trợ thanh toán thẻ trên")
		}
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	checkFeeRes, err := us.TransactionRepository.CheckTransactionQuotaAndFee(ctx, &service_transaction.CheckTransactionQuotaAndFeeReq{
		Amount:       amount,
		UserId:       request.GpayUserId,
		ServiceType:  constants.SERVICE_TYPE_WALLET,
		TransType:    request.TransType,
		SubTransType: request.SubTransType,
		SourceOfFund: sof,
		IsVerify:     true,
		IsLinkedBank: true,
		VoucherCode:  request.VoucherCode,
		ServiceCode:  request.ServiceCode,
	})

	if err != nil {
		return response, err
	}

	discountAmount = checkFeeRes.DiscountAmount
	userFee = checkFeeRes.FeeAmount

	//@todo InitOrder
	err = sg.AddStep(&saga.Step{
		Name: "INIT_ORDER",
		Func: func(c context.Context) (err error) {
			if request.OrderId != "" {
				orderDto, err = us.GetValidOrder(ctx, request.OrderId)
				if err != nil {
					return err
				}
				orderDto.SourceOfFund = sof
				orderDto.VoucherCode = request.GetVoucherCode()
				orderDto.GPayBankCode = request.GetGpayBankCode()
				orderDto.ToUserID = request.GetGpayUserId()
				orderDto.UserID = request.GetGpayUserId()
				orderDto.MerchantCode = request.GetMerchantCode()
				if orderDto.SubscribeMerchantID == "" {
					orderDto.SubscribeMerchantID = request.GetSubscriberMerchantId()
				}
				orderDto.BankCode = getLinkInfo.Data.BankCode
				orderDto.ServiceCode = request.GetServiceCode()

				orderDto.OrderCardTelco = request.GetTelco()
				orderDto.OrderBillCustomerRef = request.GetCustomerReference()
				orderDto.OrderBillServiceCode = request.GetServiceCodeBill()
				orderDto.OrderBillAreaCode = request.GetAreaCode()
				orderDto.PhoneTopUp = request.GetPhoneTopup()

			} else {
				serviceType := request.ServiceType
				if serviceType == "" {
					serviceType = constants.SERVICE_TYPE_WALLET
				}

				orderDto, err = us.InitOrder(ctx, &entities.OrderEntity{
					UserID:               request.GetGpayUserId(),
					ServiceType:          serviceType,
					OrderType:            request.GetTransType(),
					SubOrderType:         request.GetSubTransType(),
					SourceOfFund:         sof,
					Amount:               request.GetAmount(),
					VoucherCode:          request.GetVoucherCode(),
					BankCode:             getLinkInfo.Data.BankCode,
					GPayBankCode:         request.GpayBankCode,
					ToUserID:             request.GetGpayUserId(),
					OrderCardTelco:       request.GetTelco(),
					Quantity:             request.GetQuantity(),
					OrderBillServiceCode: request.GetServiceCodeBill(),
					OrderBillCustomerRef: request.GetCustomerReference(),
					OrderBillAreaCode:    request.GetAreaCode(),
					PhoneTopUp:           request.GetPhoneTopup(),
					SubscribeMerchantID:  request.GetSubscriberMerchantId(),
					MerchantCode:         request.GetMerchantCode(),
					ServiceCode:          request.GetServiceCode(),
					RefID:                request.GetRefId(),
				})
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

	//@todo InitTrans
	err = sg.AddStep(&saga.Step{
		Name: "INIT_PAYMENT",
		Func: func(c context.Context) (err error) {
			if orderDto.TransactionID != "" {
				trans, err = us.TransactionRepository.FindTransactionByID(context.TODO(), &service_transaction.ETransactionDTO{
					TransactionId: orderDto.TransactionID,
				})
			} else {
				payerId, payeeId := orderDto.UserID, orderDto.UserID
				var merchantTypeWallet, providerMerchantId string

				switch orderDto.OrderType {
				case constants.TRANSTYPE_WALLET_CASH_IN, constants.TRANSTYPE_OPENWAL_CASHIN:
					payerId = ""
				case constants.TRANSTYPE_PAY_TO_MERCHANT:
					if orderDto.SubOrderType != constants.SUB_TRANSTYPE_WALLET_APP_TO_APP {
						orderDto.SubOrderType = request.GetSubTransType()
					}
					providerMerchantId = orderDto.SubscribeMerchantID
					payeeId = ""
				case constants.TRANSTYPE_OPENWAL_CASHOUT:
					payeeId = ""
				default:
					orderDto.OrderType = constants.TRANSTYPE_WALLET_PAY_BY_TOKEN
					payeeId = ""
				}

				if orderDto.SubscribeMerchantID != "" {
					merchantTypeWallet = constants.AmountRevenue
				}

				dto := &service_transaction.ETransactionDTO{
					AppId:              orderDto.ServiceCode,
					ServiceType:        orderDto.ServiceType,
					TransactionType:    orderDto.OrderType,
					SubTransactionType: orderDto.SubOrderType,
					Amount:             amount,
					AmountFeeGpay:      userFee,
					AmountDiscount:     discountAmount,
					VoucherCode:        orderDto.VoucherCode,
					SourceOfFund:       orderDto.SourceOfFund,
					RefId:              orderDto.RefID,
					OrderId:            orderDto.OrderID,
					PayerId:            payerId,
					PayeeId:            payeeId,
					MerchantID:         orderDto.SubscribeMerchantID,
					BankCode:           orderDto.BankCode,
					MerchantTypeWallet: merchantTypeWallet,
					ProviderMerchantID: providerMerchantId,
					IbftReceiveBank:    ibftReceiveBank,
					CardNo:             getLinkInfo.Data.CardNumber,
					IbftType:           getLinkInfo.Data.Brand,
				}

				trans, err = us.serviceTransactionInit(ctx, dto)
				if err != nil {
					return err
				}

				orderDto.TransactionID = trans.TransactionId
			}

			return err
		},
		CompensateFunc: func(c context.Context) (err error) {
			if trans != nil && trans.Status != constants.TRANSACTION_STATUS_FAILED && trans.Status != constants.TRANSACTION_STATUS_PENDING {
				trans.BankTransactionId = bankTraceId
				trans.FailReason = failReason
				trans, err = us.serviceTransactionCancel(ctx, trans)
			}
			if !orderDto.Status.IsVerifying() && !orderDto.Status.IsFailed() {
				orderDto.FailReason = failReason
				us.FailedOrder(ctx, orderDto)
			}
			return
		},
	})

	//@todo PROCESSING
	err = sg.AddStep(&saga.Step{
		Name: "PROCESSING_ORDER",
		Func: func(c context.Context) (err error) {
			orderDto, err = us.ProcessingOrder(ctx, orderDto)
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			return err
		},
		Options: nil,
	})
	if err != nil {
		return
	}

	//@todo BANK GW
	err = sg.AddStep(&saga.Step{
		Name: "BANK_GW",
		Func: func(c context.Context) (err error) {
			dataReq := eBank.NapasCashInDataRequest{
				Amount:            trans.LastAmount,
				GpayTransactionID: orderDto.OrderID,
				LinkID:            request.LinkId,
				GPayUserID:        request.GpayUserId,
				Description:       "InternationalCard",
				Channel:           "WEB",
			}

			res, err := us.BankServiceRepository.CashInNapas(dataReq, request.GpayBankCode, "core.ip")
			if err != nil {
				failReason = constants.SERVICE_BANKGW_ERROR + err.Error()
				return err
			}

			bRes, err := json.Marshal(res.Data)
			if err != nil {
				panic(err)
			}
			response.Response = bRes

			type DataResponse struct {
				AcsUrl string `json:"acs_url"`
			}
			var dataResponse DataResponse
			err = json.Unmarshal(bRes, &dataResponse)

			response.Url = dataResponse.AcsUrl

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

	ordinator := saga.NewCoordinator(ctx, ctx, sg, us.LogSaga)
	rg := ordinator.Play()
	err = rg.ExecutionError
	return
}
