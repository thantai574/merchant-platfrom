package application

import (
	"context"
	"go.uber.org/zap"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	"orders-system/proto/order_system"
	"orders-system/proto/service_transaction"
	"orders-system/utils/saga"
)

func (us *OrderApplication) CashInNAPAS(ctx context.Context, request *order_system.CashInNapasRequest, response *order_system.CashInNapasResponse) (orderDto *entities.OrderEntity, err error) {
	sg := saga.NewSaga("NAPAS CASH IN ACTION")

	//todo GetOrder
	err = sg.AddStep(&saga.Step{
		Name: "INIT_ORDER",
		Func: func(c context.Context) (err error) {
			orderDto, err = us.GetValidOrder(ctx, request.OrderRequest.BankTransactionId)
			if err != nil {
				us.Logger.With(zap.Error(err)).Error("err_get_order_napas")
				return err
			}

			orderDto.BankCode = request.OrderRequest.BankCode
			return err
		},
		CompensateFunc: func(c context.Context) (err error) {
			return err
		},
		Options: nil,
	})
	if err != nil {
		return
	}

	var trans *service_transaction.ETransactionDTO
	//todo InitTrans
	err = sg.AddStep(&saga.Step{
		Name: "GET TRANSACTION ",
		Func: func(c context.Context) (err error) {
			trans, err = us.TransactionRepository.FindTransactionByID(context.TODO(), &service_transaction.ETransactionDTO{
				TransactionId: orderDto.TransactionID,
			})
			if err != nil {
				return err
			}
			trans.BankCode = request.OrderRequest.GPayBankCode
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			return err
		},
	})
	if err != nil {
		return
	}

	err = sg.AddStep(&saga.Step{
		Name: "CHECK AMOUNT ORDER WITH REQUEST",
		Func: func(c context.Context) (err error) {
			if request.OrderRequest.Amount != orderDto.Amount {
				desc := "GD nghi vấn "
				if trans.Status != constants.TRANSACTION_STATUS_PENDING {
					trans.Message = desc
					trans, err = us.serviceTransactionPending(ctx, trans)
				}
				if !orderDto.Status.IsVerifying() {
					orderDto.Description = desc
					_, _ = us.VerifyingOrder(ctx, orderDto)
				}
			}
			return err
		},
		CompensateFunc: func(c context.Context) (err error) {
			return err
		},
		Options: nil,
	})

	err = sg.AddStep(&saga.Step{
		Name: "FINALIZATION_PAYMENT",
		Func: func(c context.Context) (err error) {
			if request.ErrorCode == "200" {
				trans, err = us.serviceTransactionConfirm(ctx, trans)
				if err != nil {
					us.Logger.With().Error("SERVICE_TRANSACTION.confirm")
					return err
				}

				_, err := us.SuccessOrder(ctx, orderDto)
				return err
			} else {
				if trans != nil && trans.Status != constants.TRANSACTION_STATUS_FAILED && trans.Status != constants.TRANSACTION_STATUS_PENDING {
					trans.FailReason = request.OrderRequest.Message
					trans, err = us.serviceTransactionCancel(ctx, trans)
				}

				if !orderDto.Status.IsVerifying() && !orderDto.Status.IsFailed() {
					orderDto.FailReason = request.OrderRequest.Message
					orderDto, err := us.FailedOrder(ctx, orderDto)
					if err == nil {
						response.OrderEntity = orderDto.ConvertToProto()
					}
				}
			}

			return err
		},
		CompensateFunc: func(c context.Context) (err error) {
			return err
		},
		Options: nil,
	})

	ordinator := saga.NewCoordinator(ctx, ctx, sg, us.LogSaga)
	rg := ordinator.Play()
	err = rg.ExecutionError
	return
}
