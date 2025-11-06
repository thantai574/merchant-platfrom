package application

import (
	"context"
	"errors"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	"orders-system/proto/order_system"
	"orders-system/proto/service_transaction"
	"orders-system/utils/saga"
)

func (us *OrderApplication) CashOut(ctx context.Context, request *order_system.CashOutRequest, response *order_system.CashOutResponse) (order_dto *entities.OrderEntity, err error) {
	sg := saga.NewSaga("CashOut Action")

	var bankCodeGpay string
	var bankCode string

	order_dto = new(entities.OrderEntity)

	err = sg.AddStep(&saga.Step{
		Name: "Check VALID LinkID and User ID",
		Func: func(c context.Context) (err error) {
			getLinkIdInfo, err := us.BankServiceRepository.LinkInfo(request.LinkId)
			if err != nil {
				return err
			}

			if getLinkIdInfo.Data.GpayUserID != request.OrderRequest.UserID {
				return errors.New("INVALID UserId")
			}

			bankCode = getLinkIdInfo.Data.BankCode
			bankCodeGpay = getLinkIdInfo.Data.GpayBankCode

			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			return
		},
		Options: nil,
	})

	err = sg.AddStep(&saga.Step{
		Name: "INIT_ORDER",
		Func: func(c context.Context) (err error) {
			serviceType := request.OrderRequest.ServiceType
			if serviceType == "" {
				serviceType = constants.SERVICE_TYPE_WALLET
			}
			order_dto, err = us.InitOrder(ctx, &entities.OrderEntity{
				ServiceID:           request.OrderRequest.ServiceID,
				UserID:              request.OrderRequest.UserID,
				OrderType:           request.OrderRequest.TransType,
				SubOrderType:        request.OrderRequest.SubTransType,
				Amount:              request.OrderRequest.Amount,
				SourceOfFund:        constants.SOURCE_OF_FUND_BALANCE_WALLET,
				VoucherCode:         request.OrderRequest.VoucherCode,
				DeviceID:            request.OrderRequest.DeviceID,
				BankCode:            bankCode,
				ServiceType:         serviceType,
				SubscribeMerchantID: request.OrderRequest.GetMerchantID(),
				RefID:               request.OrderRequest.GetRefID(),
				ServiceCode:         request.OrderRequest.ServiceCode,
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
				TransactionType:    order_dto.OrderType,
				SubTransactionType: order_dto.SubOrderType,
				ServiceType:        order_dto.ServiceType,
				Amount:             order_dto.Amount,
				SourceOfFund:       order_dto.SourceOfFund,
				PayerId:            order_dto.UserID,
				AppId:              order_dto.ServiceCode,
				OrderId:            order_dto.OrderID,
				BankCode:           order_dto.BankCode,
				RefId:              order_dto.RefID,
				MerchantID:         order_dto.SubscribeMerchantID,
			}

			trans, err = us.serviceTransactionInit(ctx, dto)

			if err != nil {
				failReason = err.Error()
				return err
			}

			order_dto.TransactionID = trans.TransactionId
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
	isTimeOutBank := false
	var bankTraceId string
	err = sg.AddStep(&saga.Step{
		Name: "Cash out",
		Func: func(c context.Context) (err error) {
			cashOutResponse, err := us.BankServiceRepository.CashOut(request.OrderRequest.Amount, bankCodeGpay,
				order_dto.OrderID, request.LinkId, request.Description, request.OrderRequest.UserID)
			if err != nil {
				bankTraceId = cashOutResponse.Data.BankTraceId
				failReason = err.Error()
				return err
			}

			bankTraceId = cashOutResponse.Data.BankTraceId

			if cashOutResponse.ErrorCode.IsVerifying() {
				isTimeOutBank = true
				return errors.New(cashOutResponse.Message)
			}

			bankTraceId = cashOutResponse.Data.BankTraceId
			order_dto.BankTransactionId = bankTraceId
			if order_dto.BankTransactionId == "" {
				order_dto.BankTransactionId = order_dto.OrderID
			}
			return err
		},
		CompensateFunc: func(c context.Context) (err error) {
			if isTimeOutBank == true {
				if trans.Status != constants.TRANSACTION_STATUS_PENDING {
					trans.BankTransactionId = bankTraceId
					trans, err = us.serviceTransactionPending(ctx, trans)
				}
				if !order_dto.Status.IsVerifying() {
					us.VerifyingOrder(ctx, order_dto)
				}
			} else {
				trans.BankTransactionId = bankTraceId
				trans.FailReason = failReason
				trans, err = us.serviceTransactionCancel(ctx, trans)

				order_dto.FailReason = failReason
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
			trans.BankTransactionId = order_dto.BankTransactionId
			trans, err = us.serviceTransactionConfirm(ctx, trans)
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			if trans.Status != constants.TRANSACTION_STATUS_PENDING {
				trans.BankTransactionId = bankTraceId

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
