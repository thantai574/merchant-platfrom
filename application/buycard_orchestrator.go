package application

import (
	"context"
	"fmt"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	"orders-system/errors"
	"orders-system/proto/order_system"
	"orders-system/proto/service_card"
	"orders-system/proto/service_merchant_fee"
	"orders-system/proto/service_promotion"
	"orders-system/proto/service_transaction"
	mapErrString "orders-system/utils/errors"
	"orders-system/utils/saga"
	"strings"

	"go.uber.org/zap"
)

func (us *OrderApplication) BuyCard(ctx context.Context, request *order_system.OrderBuyCardRequest, response *order_system.OrderBuyCardResponse) (orderDto *entities.OrderEntity, err error) {
	sg := saga.NewSaga("BuyCardActionAccount")

	discount := int64(0)

	voucherId := ""
	var failReason string

	var paymentTokenOrder entities.OrderEntity

	if request.OrderRequest.SourceOfFund == constants.SOURCE_OF_FUND_BANK_ATM {
		getPaymentTokenOrder, err := us.GetValidOrder(ctx, request.ConfirmPaymentTokenRequest.OrderId)
		if err != nil {
			return nil, err
		}
		paymentTokenOrder = *getPaymentTokenOrder
	}

	orderDto = &entities.OrderEntity{
		OrderID:             paymentTokenOrder.OrderID,
		ServiceID:           request.OrderRequest.ServiceID,
		UserID:              request.OrderRequest.UserID,
		SubscribeMerchantID: request.OrderRequest.MerchantID,
		OrderType:           request.OrderRequest.TransType,
		SubOrderType:        request.OrderRequest.SubTransType,
		Amount:              request.OrderRequest.Amount,
		Quantity:            request.OrderRequest.Quantity,
		SourceOfFund:        request.OrderRequest.SourceOfFund,
		VoucherCode:         request.OrderRequest.VoucherCode,
		DeviceID:            request.OrderRequest.DeviceID,
		PhoneTopUp:          request.OrderRequest.PhoneTopup,
		CreatedAt:           paymentTokenOrder.CreatedAt,
		UpdatedAt:           paymentTokenOrder.UpdatedAt,
		PaymentOrderId:      request.LongLifeOrderId,
	}

	if orderDto.OrderID == "" {
		err = sg.AddStep(&saga.Step{
			Name: "INIT_ORDER",
			Func: func(c context.Context) (err error) {
				orderDto, err = us.InitOrder(ctx, orderDto)

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
			return
		}
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
		return
	}

	var cardNameMerchantQuotaCheck string
	if strings.Contains(request.Telco, "DT_") {
		cardNameMerchantQuotaCheck = strings.Replace(request.Telco, "DT_", "", -1)
	} else {
		cardNameMerchantQuotaCheck = request.Telco
	}

	err = sg.AddStep(&saga.Step{
		Name: "ACCOUNT_VOUCHER",
		Func: func(c context.Context) (err error) {
			if request.OrderRequest.VoucherCode != "" {
				res, err := us.servicePromotionUsed(ctx, &service_promotion.UseVoucherRequest{
					Code:         orderDto.VoucherCode,
					UserId:       orderDto.UserID,
					TraceId:      orderDto.OrderID,
					Total:        1,
					Amount:       request.OrderRequest.Quantity * request.OrderRequest.Amount,
					ServiceCode:  request.OrderRequest.SubTransType,
					SourceOfFund: request.OrderRequest.SourceOfFund,
				})
				if err != nil {
					return err
				}
				voucherId = res.Voucher.Voucher.Id
				discount = res.DiscountAmount
			}
			return err
		},
		CompensateFunc: func(c context.Context) (err error) {
			_, err = us.servicePromotionCompensate(ctx, &service_promotion.ReverseWalletRequest{
				TraceId: orderDto.OrderID,
			})
			if !orderDto.Status.IsVerifying() && !orderDto.Status.IsFailed() {
				_, _ = us.FailedOrder(ctx, orderDto)
			}
			return
		},
		Options: nil,
	})
	if err != nil {
		return
	}

	var trans *service_transaction.ETransactionDTO

	err = sg.AddStep(&saga.Step{
		Name: "INIT_PAYMENT",
		Func: func(c context.Context) (err error) {
			dto := &service_transaction.ETransactionDTO{
				TransactionType:    request.OrderRequest.TransType,
				SubTransactionType: request.OrderRequest.SubTransType,
				ServiceType:        constants.SERVICE_TYPE_WALLET,
				Amount:             request.OrderRequest.Amount * request.OrderRequest.Quantity,
				DeviceId:           request.OrderRequest.DeviceID,
				LastAmount:         request.OrderRequest.Amount * request.OrderRequest.Quantity,
				VoucherCode:        request.OrderRequest.VoucherCode,
				AmountDiscount:     discount,
				SourceOfFund:       request.OrderRequest.SourceOfFund,
				PayerId:            request.OrderRequest.UserID,
				GpayAccountID:      constants.GPAY_ACCOUNT_ID,
				GpayTypeWallet:     constants.Amount,
				AppId:              request.OrderRequest.SubTransType,
				VoucherID:          voucherId,
				OrderId:            orderDto.OrderID,
				RefId:              request.OrderRequest.RefID,
			}

			if request.OrderRequest.MerchantID != "" {
				dto.SubscriberMerchantID = request.OrderRequest.MerchantID
				dto.SubscriberMerchantTypeWallet = constants.AmountCash
			}

			trans, err = us.serviceTransactionInit(ctx, dto)

			if err != nil {
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
				_, _ = us.FailedOrder(ctx, orderDto)
			}
			return
		},
	})
	if err != nil {
		return
	}

	var providerMerchantId, providerMerchant, cardType string

	isPendingSrvCard := false
	// call card service
	err = sg.AddStep(&saga.Step{
		Name: "BUY_CARD",
		Func: func(c context.Context) (err error) {
			buycardRes, err := us.ServiceCardRepository.BuyCard(ctx, &service_card.BuyCardReq{
				Price:         request.OrderRequest.Amount,
				Telco:         request.Telco,
				Quantity:      request.OrderRequest.Quantity,
				UserId:        request.OrderRequest.UserID,
				OrderId:       orderDto.OrderID,
				TransactionId: trans.TransactionId,
			})
			response.Cards = []*order_system.CardObjDTO{}
			if err != nil {
				failReason = mapErrString.GetGrpcErrMessage(err)
				return fmt.Errorf("%v", failReason)
			}
			if buycardRes.Cards != nil {
				for _, v := range buycardRes.Cards {
					response.Cards = append(response.Cards, &order_system.CardObjDTO{
						Provider:    v.Provider,
						CardNumber:  v.CardNumber,
						Serial:      v.Serial,
						Price:       v.Price,
						NameCard:    v.NameCard,
						PackageData: v.PackageData,
						Period:      v.Period,
					})

				}
			}
			providerMerchantId = buycardRes.ProviderMerchantId
			providerMerchant = buycardRes.ProviderMerchant
			cardType = buycardRes.CardType

			if buycardRes != nil && buycardRes.Status == constants.TRANSACTION_STATUS_PENDING {
				isPendingSrvCard = true
				return errors.ErrPendingOrder
			}

			return err
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
			return
		},
	})
	if err != nil {
		return
	}
	//

	//todo Accounting Merchant Fee
	err = sg.AddStep(&saga.Step{
		Name: "PROVIDER CARD FEE",
		Func: func(c context.Context) (err error) {
			merchantFee, err := us.MerchantFeeRepository.GetMerchantVendorDiscount(ctx, &service_merchant_fee.GetMerchantVendorDiscountReq{
				Amount:       request.OrderRequest.Amount,
				ServiceType:  trans.ServiceType,
				TransType:    trans.TransactionType,
				SubTransType: trans.SubTransactionType,
				MerchantId:   request.OrderRequest.MerchantID,
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
				trans.DiscountForMerchantSubscriber = request.OrderRequest.Quantity * merchantFee.MerchantDiscountAmount // chiet khau gpay cho merchant
				trans.DiscountOfMerchantProvider = request.OrderRequest.Quantity * merchantFee.VendorDiscountAmount      // chiet khau provider ZOTA/IMEDIA cho GPAY
			}
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			return
		},
	})
	if err != nil {
		return
	}

	//todo Confirm Trans
	err = sg.AddStep(&saga.Step{
		Name: "FINALIZATION_PAYMENT",
		Func: func(c context.Context) (err error) {
			trans.ProviderMerchantID = providerMerchantId
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
		return
	}

	//todo Confirm Order
	err = sg.AddStep(&saga.Step{
		Name: "SUCCESS_ORDER",
		Func: func(c context.Context) (err error) {
			orderDto.TransactionID = trans.TransactionId
			orderDto, err = us.SuccessOrder(ctx, orderDto)
			if err == nil {
				response.OrderEntity = orderDto.ConvertToProto()
			}
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			if trans.Status != constants.TRANSACTION_STATUS_PENDING {
				trans, err = us.serviceTransactionPending(ctx, trans)
			}
			if !orderDto.Status.IsVerifying() {
				us.VerifyingOrder(ctx, orderDto)
			}
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
