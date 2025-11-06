package application

import (
	"context"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	"orders-system/proto/order_system"
	"orders-system/proto/service_transaction"
	"orders-system/utils/saga"
)

func (us *OrderApplication) TransferGpayWallet(ctx context.Context, request *order_system.TransferWalletRequest) (response *order_system.TransferWalletResponse, err error) {
	sg := saga.NewSaga("TransferGpayWallet")

	response = new(order_system.TransferWalletResponse)
	var orderDto *entities.OrderEntity

	err = sg.AddStep(&saga.Step{
		Name: "INIT_ORDER",
		Func: func(c context.Context) (err error) {
			orderDto, err = us.InitOrder(ctx, &entities.OrderEntity{
				UserID:              request.OrderRequest.GetUserID(),
				SubscribeMerchantID: request.OrderRequest.GetMerchantID(),
				OrderType:           request.OrderRequest.GetTransType(),
				Amount:              request.OrderRequest.GetAmount(),
				SourceOfFund:        constants.SOURCE_OF_FUND_BALANCE_WALLET,
				BankCode:            request.OrderRequest.GetBankCode(),
				ToUserID:            request.OrderRequest.GetToUserID(),
				Description:         request.OrderRequest.GetDescription(),
				ServiceCode:         constants.TRANSTYPE_WALLET_TRANSFER,
				ServiceType:         request.OrderRequest.ServiceType,
			})
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

	// PROCESSING ORDER
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

	var failReason string

	//INIT CORE V3
	var trans *service_transaction.ETransactionDTO

	err = sg.AddStep(&saga.Step{
		Name: "INIT_PAYMENT",
		Func: func(c context.Context) (err error) {
			dto := &service_transaction.ETransactionDTO{
				TransactionType:    request.OrderRequest.TransType,
				SubTransactionType: request.OrderRequest.SubTransType,
				SourceOfFund:       request.OrderRequest.SourceOfFund,
				ServiceType:        constants.SERVICE_TYPE_WALLET,
				PayerId:            request.OrderRequest.UserID,
				PayeeId:            request.OrderRequest.ToUserID,
				Message:            request.OrderRequest.Message,
				Amount:             request.OrderRequest.Amount,
				AppId:              request.OrderRequest.ServiceCode,
				MerchantID:         request.OrderRequest.MerchantID,
				RefId:              request.OrderRequest.RefID,
			}

			trans, err = us.serviceTransactionInit(ctx, dto)

			if err != nil {
				failReason = err.Error()
				return err
			}
			orderDto.TransactionID = trans.TransactionId
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

	//todo Confirm Trans
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
		return
	}

	err = sg.AddStep(&saga.Step{
		Name: "SUCCESS_ORDER",
		Func: func(c context.Context) (err error) {
			orderDto, err = us.SuccessOrder(ctx, orderDto)
			if err == nil {
				response.OrderEntity = orderDto.ConvertToProto()
			}
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			if !orderDto.Status.IsVerifying() {
				_, _ = us.VerifyingOrder(ctx, orderDto)
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
