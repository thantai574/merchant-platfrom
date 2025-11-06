package application

import (
	"context"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	"orders-system/proto/order_system"
	"orders-system/proto/service_promotion"
	"orders-system/proto/service_transaction"
	"orders-system/utils/saga"
)

func (us *OrderApplication) StaticQR(ctx context.Context, request *order_system.PaymentMerchantRequest, response *order_system.PaymentMerchantResponse) (order_dto *entities.OrderEntity, err error) {
	sg := saga.NewSaga("Payment Merchant")

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
				SourceOfFund:        request.OrderRequest.SourceOfFund,
				VoucherCode:         request.OrderRequest.VoucherCode,
				DeviceID:            request.OrderRequest.DeviceID,
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
				SubTransactionType: request.OrderRequest.SubTransType,
				ServiceType:        constants.SERVICE_TYPE_WALLET,
				Amount:             request.OrderRequest.Amount,
				DeviceId:           request.OrderRequest.DeviceID,
				LastAmount:         request.OrderRequest.Amount,
				VoucherCode:        request.OrderRequest.VoucherCode,
				AmountDiscount:     discount,
				SourceOfFund:       request.OrderRequest.SourceOfFund,
				PayerId:            request.OrderRequest.UserID,
				AppId:              request.OrderRequest.SubTransType,
				VoucherID:          voucherId,
				OrderId:            order_dto.OrderID,
				MerchantTypeWallet: constants.AmountRevenue,
				Message:            request.Message,
				MerchantID:         request.OrderRequest.MerchantID,
			}

			if request.OrderRequest.MerchantID != "" {
				dto.ProviderMerchantID = request.OrderRequest.MerchantID
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
				us.FailedOrder(ctx, order_dto)
			}
			return
		},
	})

	if err != nil {
		return
	}

	err = sg.AddStep(&saga.Step{
		Name: "FINALIZATION_PAYMENT",
		Func: func(c context.Context) (err error) {
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
