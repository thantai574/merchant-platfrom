package application

import (
	"context"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	"orders-system/proto/order_system"
	"orders-system/proto/service_transaction"
	"orders-system/utils/saga"
)

func (us *OrderApplication) LixiOrchestrator(ctx context.Context, request *order_system.LixiRequest, res *order_system.LixiResponse) (err error) {
	sg := saga.NewSaga("LixiOrchestrator")

	var order_dto *entities.OrderEntity
	var trans *service_transaction.ETransactionDTO

	err = sg.AddStep(&saga.Step{
		Name: "Init-Order",
		Func: func(c context.Context) (err error) {
			order_dto, err = us.InitOrder(ctx, &entities.OrderEntity{
				ServiceID:           constants.TRANSTYPE_WALLET_LIXI,
				UserID:              request.OrderRequest.ToUserID,
				SubscribeMerchantID: request.OrderRequest.MerchantID,
				OrderType:           constants.TRANSTYPE_WALLET_LIXI,
				SubOrderType:        constants.TRANSTYPE_WALLET_LIXI,
				Amount:              request.OrderRequest.Amount,
				SourceOfFund:        request.OrderRequest.SourceOfFund,
				VoucherCode:         request.OrderRequest.VoucherCode,
				DeviceID:            request.OrderRequest.DeviceID,
				ToUserID:            request.OrderRequest.ToUserID,
				RefID:               request.Lixi.ID,
				LuckyMoneyID:        request.Lixi.ID,
			})
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			return
		},
		Options: nil,
	})

	err = sg.AddStep(&saga.Step{
		Name: "Processing-Order",
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

	err = sg.AddStep(&saga.Step{
		Name: "Init-Transaction",
		Func: func(c context.Context) (err error) {
			dto := &service_transaction.ETransactionDTO{
				TransactionType: constants.TRANSTYPE_WALLET_LIXI,
				ServiceType:     constants.SERVICE_TYPE_WALLET,
				Amount:          request.OrderRequest.Amount,
				DeviceId:        request.OrderRequest.DeviceID,
				LastAmount:      request.OrderRequest.Amount,
				SourceOfFund:    constants.SOURCE_OF_FUND_BALANCE_WALLET,
				PayerId:         request.OrderRequest.UserID,
				GpayAccountID:   constants.GPAY_ACCOUNT_ID,
				GpayTypeWallet:  constants.Amount,
				AppId:           constants.TRANSTYPE_WALLET_LIXI,
				OrderId:         order_dto.OrderID,
				LuckyMoneyID:    request.Lixi.ID,
				PayeeId:         request.OrderRequest.ToUserID,
				Message:         request.Lixi.Name,
			}

			if request.OrderRequest.MerchantID != "" {
				dto.MerchantID = request.OrderRequest.MerchantID
				dto.MerchantTypeWallet = constants.AmountCash
			}

			trans, err = us.serviceTransactionInit(ctx, dto)
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
		Options: nil,
	})
	// t
	err = sg.AddStep(&saga.Step{
		Name: "Confirm-Transaction",
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

	err = sg.AddStep(&saga.Step{
		Name: "Success-Order",
		Func: func(c context.Context) (err error) {
			order_dto, err = us.SuccessOrder(ctx, order_dto)
			if err == nil {
				res.Order = order_dto.ConvertToProto()
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
	})
	coordinator := saga.NewCoordinator(ctx, ctx, sg, us.LogSaga)
	rg := coordinator.Play()
	err = rg.ExecutionError

	return
}
