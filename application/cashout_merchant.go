package application

import (
	"context"
	"errors"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	"orders-system/proto/order_system"
	"orders-system/proto/service_transaction"
	mapErrString "orders-system/utils/errors"
	"orders-system/utils/saga"
)

// rut tien merchant
func (us *OrderApplication) WithdrawMerchant(ctx context.Context, request *order_system.WithrawMerchantRequest) (order_dto *entities.OrderEntity, err error) {
	sg := saga.NewSaga("Withdraw Merchant")

	err = sg.AddStep(&saga.Step{
		Name: "INIT_ORDER",
		Func: func(c context.Context) (err error) {
			order_dto, err = us.InitOrder(ctx, &entities.OrderEntity{
				ServiceID:    constants.TRANSTYPE_WALLET_CASH_OUT_MC,
				ServiceType:  constants.SERVICE_TYPE_WALLET,
				OrderType:    constants.TRANSTYPE_WALLET_CASH_OUT_MC,
				Amount:       request.OrderRequest.Amount,
				SourceOfFund: constants.SOURCE_OF_FUND_BALANCE_WALLET,
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

	var trans *service_transaction.ETransactionDTO
	var failReason string

	err = sg.AddStep(&saga.Step{
		Name: "INIT_PAYMENT",
		Func: func(c context.Context) (err error) {
			dto := &service_transaction.ETransactionDTO{
				TransactionType:    constants.TRANSTYPE_WALLET_CASH_OUT_MC,
				ServiceType:        constants.SERVICE_TYPE_WALLET,
				Amount:             request.OrderRequest.Amount,
				LastAmount:         request.OrderRequest.Amount,
				TypeWallet:         constants.AmountCash,
				SourceOfFund:       constants.SOURCE_OF_FUND_BALANCE_WALLET,
				GpayAccountID:      constants.GPAY_ACCOUNT_ID,
				GpayTypeWallet:     constants.Amount,
				OrderId:            order_dto.OrderID,
				Message:            request.Message,
				MerchantTypeWallet: constants.AmountCash,
				BankCode:           request.OrderRequest.BankCode,
				AdminId:            request.AdminId,
				BankTypeWallet:     request.BankAccountType,
				MerchantID:         request.OrderRequest.MerchantID,
			}

			trans, err = us.serviceTransactionInit(ctx, dto)
			if err != nil {
				failReason = mapErrString.GetGrpcErrMessage(err)
				return errors.New(failReason)
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
