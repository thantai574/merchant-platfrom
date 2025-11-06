package application

import (
	"context"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	errorsMap "orders-system/errors"
	"orders-system/proto/order_system"
	"orders-system/proto/service_card"
	"orders-system/proto/service_promotion"
	"orders-system/proto/service_transaction"
	"orders-system/utils/saga"
)

func (us *OrderApplication) TopUp(ctx context.Context, request *order_system.OrderTopUpRequest, response *order_system.OrderTopUpResponse) (order_dto *entities.OrderEntity, err error) {
	sg := saga.NewSaga("TopUpActionAccount")

	discount := int64(0)
	voucherId := ""

	err = sg.AddStep(&saga.Step{
		Name: "INIT_ORDER",
		Func: func(c context.Context) (err error) {
			order_dto, err = us.InitOrder(ctx, &entities.OrderEntity{
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
				MerchantCode:        request.OrderRequest.MerchantCode,
			})
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			if !order_dto.Status.IsVerifying() && !order_dto.Status.IsFailed() {
				us.FailedOrder(ctx, order_dto)
			}
			return
		},
		Options: nil,
	})
	if err != nil {
		return
	}

	err = sg.AddStep(&saga.Step{
		Name: "PROCESSING_ORDER",
		Func: func(c context.Context) (err error) {
			order_dto, err = us.ProcessingOrder(ctx, order_dto)
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			if !order_dto.Status.IsVerifying() && !order_dto.Status.IsFailed() {
				us.FailedOrder(ctx, order_dto)
			}
			return
		},
		Options: nil,
	})
	if err != nil {
		return
	}

	err = sg.AddStep(&saga.Step{
		Name: "ACCOUNT_VOUCHER",
		Func: func(c context.Context) (err error) {
			if request.OrderRequest.VoucherCode != "" {
				res, err := us.servicePromotionUsed(ctx, &service_promotion.UseVoucherRequest{
					Code:         order_dto.VoucherCode,
					UserId:       order_dto.UserID,
					TraceId:      order_dto.OrderID,
					Total:        1,
					Amount:       request.OrderRequest.Amount,
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
				TraceId: order_dto.OrderID,
			})
			if !order_dto.Status.IsVerifying() && !order_dto.Status.IsFailed() {
				us.FailedOrder(ctx, order_dto)
			}
			return
		},
		Options: nil,
	})
	if err != nil {
		return
	}

	var trans *service_transaction.ETransactionDTO
	var failReason string

	err = sg.AddStep(&saga.Step{
		Name: "INIT_PAYMENT",
		Func: func(c context.Context) (err error) {
			dto := &service_transaction.ETransactionDTO{
				TransactionType:    request.OrderRequest.TransType,
				SubTransactionType: constants.SUB_TRANSTYPE_WALLET_TOPUP_CARD,
				ServiceType:        constants.SERVICE_TYPE_WALLET,
				Amount:             request.OrderRequest.Amount,
				DeviceId:           request.OrderRequest.DeviceID,
				LastAmount:         request.OrderRequest.Amount,
				VoucherCode:        request.OrderRequest.VoucherCode,
				AmountDiscount:     discount,
				SourceOfFund:       request.OrderRequest.SourceOfFund,
				PayerId:            request.OrderRequest.UserID,
				GpayAccountID:      constants.GPAY_ACCOUNT_ID,
				GpayTypeWallet:     constants.Amount,
				AppId:              constants.SUB_TRANSTYPE_WALLET_TOPUP_CARD,
				VoucherID:          voucherId,
				OrderId:            order_dto.OrderID,
				RefId:              request.OrderRequest.RefID,
			}

			if request.OrderRequest.MerchantID != "" {
				dto.SubscriberMerchantID = request.OrderRequest.MerchantID
				dto.SubscriberMerchantTypeWallet = constants.AmountCash
			}

			trans, err = us.serviceTransactionInit(ctx, dto)
			if err != nil {
				failReason = err.Error()
			}
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			if trans != nil && trans.Status != constants.TRANSACTION_STATUS_FAILED && trans.Status != constants.TRANSACTION_STATUS_PENDING {
				trans.FailReason = failReason
				trans, err = us.serviceTransactionCancel(ctx, trans)
			}
			if !order_dto.Status.IsVerifying() && !order_dto.Status.IsFailed() {
				order_dto.InternalErr = failReason
				us.FailedOrder(ctx, order_dto)
			}
			return
		},
	})
	if err != nil {
		return
	}

	// call card service
	var providerMerchantId string
	isPendingSrvCard := false
	err = sg.AddStep(&saga.Step{
		Name: "TOP_UP",
		Func: func(c context.Context) (err error) {
			topUpRes, err := us.ServiceCardRepository.Topup(ctx, &service_card.TopupReq{
				Price:         request.OrderRequest.Amount,
				Telco:         request.Telco,
				Phone:         request.OrderRequest.PhoneTopup,
				UserId:        request.OrderRequest.UserID,
				OrderId:       order_dto.OrderID,
				SubTransType:  request.OrderRequest.SubTransType,
				TransactionId: trans.TransactionId,
			})

			if err != nil {
				failReason = err.Error()
				return errorsMap.ErrFailOrder
			}

			providerMerchantId = topUpRes.ProviderMerchantId

			if topUpRes != nil && topUpRes.Status == constants.TRANSACTION_STATUS_PENDING {
				isPendingSrvCard = true
				return errorsMap.ErrPendingOrder
			}

			return err
		},
		CompensateFunc: func(c context.Context) (err error) {
			if isPendingSrvCard == true { // Pending Order Service Card
				if trans.Status != constants.TRANSACTION_STATUS_PENDING {
					trans, err = us.serviceTransactionPending(ctx, trans)
				}
				if !order_dto.Status.IsVerifying() {
					us.VerifyingOrder(ctx, order_dto)
				}
			} else {
				trans.FailReason = failReason
				trans, err = us.serviceTransactionCancel(ctx, trans)

				order_dto.InternalErr = failReason
				_, err = us.FailedOrder(ctx, order_dto)
			}

			return
		},
	})
	if err != nil {
		return
	}

	//
	err = sg.AddStep(&saga.Step{
		Name: "FINALIZATION_PAYMENT",
		Func: func(c context.Context) (err error) {
			trans.ProviderMerchantID = providerMerchantId
			trans, err = us.serviceTransactionConfirm(ctx, trans)
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			if trans != nil && trans.Status != constants.TRANSACTION_STATUS_PENDING {
				trans, err = us.serviceTransactionPending(ctx, trans)
			}
			if !order_dto.Status.IsVerifying() {
				us.VerifyingOrder(ctx, order_dto)
			}

			return
		},
	})

	if err != nil {
		return
	}

	err = sg.AddStep(&saga.Step{
		Name: "SUCCESS_ORDER",
		Func: func(c context.Context) (err error) {
			order_dto, err = us.SuccessOrder(ctx, order_dto)
			if err == nil {
				response.OrderEntity = order_dto.ConvertToProto()
			}
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			if trans.Status != constants.TRANSACTION_STATUS_PENDING {
				trans, err = us.serviceTransactionPending(ctx, trans)
			}
			if !order_dto.Status.IsVerifying() {
				us.VerifyingOrder(ctx, order_dto)
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
