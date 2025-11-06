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

func (us *OrderApplication) VerifyOTP(ctx context.Context, request *order_system.OrderCashOTPRequest, response *order_system.OrderCashOTPResponse) (order_dto *entities.OrderEntity, err error) {
	sg := saga.NewSaga("VERIFY OTP ACTION")
	order_dto = new(entities.OrderEntity)

	err = sg.AddStep(&saga.Step{
		Name: "FIND_ORDER_BY_ID",
		Func: func(c context.Context) (err error) {
			findOrder, err := us.GetValidOrder(ctx, request.OrderId)
			if err != nil {
				return err
			}
			order_dto = findOrder
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			return err
		},
	})

	if err != nil {
		return
	}

	var trans *service_transaction.ETransactionDTO

	err = sg.AddStep(&saga.Step{
		Name: "GET INIT TRANSACTION",
		Func: func(c context.Context) (err error) {
			findTranById, err := us.TransactionRepository.FindTransactionByID(ctx, &service_transaction.ETransactionDTO{TransactionId: order_dto.TransactionID})
			if err != nil {
				return err
			}
			trans = findTranById
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			return err
		},
	})

	if err != nil {
		return
	}

	// call card service
	isTimeOutBank := false
	var isFailedOrder bool
	var bankTransactionId, failReason string

	err = sg.AddStep(&saga.Step{
		Name: "VERIFY OTP",
		Func: func(c context.Context) (err error) {
			cashInVerifyOTP, err := us.BankServiceRepository.VerifyOTP(order_dto.GPayBankCode, request.RefBankTraceID, request.OrderId, request.LinkId, request.Otp)
			if err != nil {
				isFailedOrder = true
				bankTransactionId = cashInVerifyOTP.Data.BankTraceID
				failReason = err.Error()
				bankTransactionId = cashInVerifyOTP.Data.BankTraceID
				return err
			}

			bankTransactionId = cashInVerifyOTP.Data.BankTraceID

			if cashInVerifyOTP.ErrorCode.IsVerifying() {
				isTimeOutBank = true
				return errors.New(cashInVerifyOTP.Message)
			}

			if cashInVerifyOTP.ErrorCode.IsWrongOTP() {
				isFailedOrder = false
				return errors.New(cashInVerifyOTP.Message)
			}

			return err
		},
		CompensateFunc: func(c context.Context) (err error) {
			if isTimeOutBank == true {
				if !order_dto.Status.IsVerifying() {
					us.VerifyingOrder(ctx, order_dto)
				}
				if trans.Status != constants.TRANSACTION_STATUS_PENDING {
					trans.BankTransactionId = bankTransactionId
					trans, err = us.serviceTransactionPending(ctx, trans)
				}
			} else {
				if isFailedOrder == true {
					trans.BankTransactionId = bankTransactionId
					trans.FailReason = failReason
					if trans != nil && trans.Status != constants.TRANSACTION_STATUS_PENDING && trans.Status != constants.TRANSACTION_STATUS_FINISH &&
						trans.Status != constants.TRANSACTION_STATUS_FAILED {
						trans, err = us.serviceTransactionCancel(ctx, trans)

						order_dto.InternalErr = failReason
						us.FailedOrder(ctx, order_dto)
					}
				}
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
			trans.BankTransactionId = bankTransactionId
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
			order_dto.TransactionID = trans.TransactionId
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
