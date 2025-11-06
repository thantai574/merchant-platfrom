package application

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	"orders-system/proto/order_system"
	"orders-system/proto/service_promotion"
	"orders-system/proto/service_transaction"
	"orders-system/utils/saga"
)

func (us *OrderApplication) FundWal2Wal(ctx context.Context, request *order_system.FundWallet2WalletRequest, response *order_system.FundWallet2WalletResponse) (order_dto *entities.OrderEntity, err error) {
	sg := saga.NewSaga("Fund WAL2WAL")
	us.Logger.With(zap.Field{
		Key:       "request Fund WAL2WAL",
		Type:      zapcore.StringType,
		Integer:   0,
		String:    fmt.Sprintf("%v", request),
		Interface: nil,
	}).Info("request")
	var discount = int64(0)
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
					Code:        order_dto.VoucherCode,
					UserId:      order_dto.UserID,
					TraceId:     order_dto.OrderID,
					Total:       1,
					Amount:      request.OrderRequest.Amount,
					ServiceCode: request.OrderRequest.ServiceID,
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
	_ = discount

	var trans *service_transaction.ETransactionDTO

	err = sg.AddStep(&saga.Step{
		Name: "INIT_PAYMENT",
		Func: func(c context.Context) (err error) {
			dto := &service_transaction.ETransactionDTO{
				TransactionType: constants.TRANSTYPE_BEHALF_PAY_WALLET_TO_WALLET,
				ServiceType:     constants.SERVICE_TYPE_COLLECTION_AND_PAY,
				Amount:          request.OrderRequest.Amount,
				DeviceId:        request.OrderRequest.DeviceID,
				LastAmount:      request.OrderRequest.Amount,
				VoucherCode:     request.OrderRequest.VoucherCode,
				SourceOfFund:    constants.SOURCE_OF_FUND_BALANCE_WALLET,
				GpayAccountID:   constants.GPAY_ACCOUNT_ID,
				TypeWallet:      constants.AmountCash,
				AppId:           constants.TRANSTYPE_BEHALF_PAY_WALLET_TO_WALLET,
				VoucherID:       voucherId,
				OrderId:         order_dto.OrderID,
				RefId:           request.OrderRequest.RefID,
				Message:         request.Description,
				PayeeId:         request.OrderRequest.UserID,
			}

			if request.OrderRequest.MerchantID != "" {
				dto.MerchantID = request.OrderRequest.MerchantID
				dto.MerchantTypeWallet = constants.AmountCash
			}

			trans, err = us.serviceTransactionInit(ctx, dto)
			if err != nil {
				return err
			}

			order_dto.TransactionID = trans.TransactionId
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			if trans != nil && trans.Status != constants.TRANSACTION_STATUS_FAILED && trans.Status != constants.TRANSACTION_STATUS_PENDING {
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

	//
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
