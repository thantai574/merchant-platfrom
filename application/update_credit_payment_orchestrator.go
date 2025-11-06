package application

import (
	"context"
	"go.uber.org/zap"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	eBankGw "orders-system/domain/entities/bank_gateway"
	"orders-system/proto/order_system"
	"orders-system/proto/service_transaction"
	"orders-system/utils/helpers"
	"orders-system/utils/saga"
)

func (us *OrderApplication) UpdateCreditPaymentOrder(ctx context.Context, request *order_system.UpdateCreditPaymentRequest, response *order_system.UpdateCreditPaymentResponse) (order_dto *entities.OrderEntity, err error) {
	sg := saga.NewSaga("Update final credit payment Order")

	order_dto, err = us.GetValidOrder(ctx, request.OrderId)
	if err != nil {
		us.Logger.Info("get_order_by_id_fail", zap.Error(err))
		return order_dto, err
	}

	var trans *service_transaction.ETransactionDTO

	findTranById, err := us.TransactionRepository.FindTransactionByID(ctx, &service_transaction.ETransactionDTO{TransactionId: order_dto.TransactionID})
	if err != nil {
		return order_dto, err
	}
	trans = findTranById

	trans.BankTransactionId = order_dto.OrderID
	trans.OrderId = order_dto.OrderID

	switch request.Status {
	case constants.STATUS_SUCCESS:
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
			return order_dto, err
		}

		err = sg.AddStep(&saga.Step{
			Name: "SUCCESS_ORDER",
			Func: func(c context.Context) (err error) {
				order_dto.SucceedAt = helpers.GetCurrentTime()
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
			return order_dto, err
		}

	case constants.STATUS_FAILED:
		if trans != nil && trans.Status != constants.TRANSACTION_STATUS_FAILED && trans.Status != constants.TRANSACTION_STATUS_PENDING {

			trans.FailReason = request.Description
			trans, err = us.serviceTransactionCancel(ctx, trans)
		}

		if !order_dto.Status.IsVerifying() && !order_dto.Status.IsFailed() {
			order_dto.FailReason = request.Description
			orderDto, err := us.FailedOrder(ctx, order_dto)
			if err == nil {
				response.OrderEntity = orderDto.ConvertToProto()
			}
		}

	case constants.STATUS_VERIFYING:
		retrieveOrderStatus, err := us.BankServiceRepository.RetrieveOrderStatus(request.OrderId, eBankGw.GpayBankCodeCreditPayment)
		if err != nil || !retrieveOrderStatus.ErrorCode.IsSuccess() {
			trans, err = us.serviceTransactionPending(ctx, trans)
			orderDto, err := us.VerifyingOrder(ctx, order_dto)

			if err == nil {
				response.OrderEntity = orderDto.ConvertToProto()
			}
			return orderDto, err
		}

		if retrieveOrderStatus.ErrorCode.IsSuccess() {
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
				return order_dto, err
			}

			err = sg.AddStep(&saga.Step{
				Name: "SUCCESS_ORDER",
				Func: func(c context.Context) (err error) {
					order_dto.SucceedAt = helpers.GetCurrentTime()
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
				return order_dto, err
			}
		}

	}

	ordinator := saga.NewCoordinator(ctx, ctx, sg, us.LogSaga)
	rg := ordinator.Play()
	err = rg.ExecutionError

	return order_dto, err
}
