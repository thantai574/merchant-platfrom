package application

import (
	"context"
	"errors"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	"orders-system/proto/order_system"
	"orders-system/proto/service_promotion"
	"orders-system/proto/service_transaction"
	"orders-system/utils/saga"
)

func (us *OrderApplication) IBFTransfer(ctx context.Context, request *order_system.IBFTTransferRequest, response *order_system.IBFTTransferResponse) (order_dto *entities.OrderEntity, err error) {
	sg := saga.NewSaga("IBFT TRANSFER Action")
	var discount int64 = 0

	voucherId := ""
	var failReason string

	err = sg.AddStep(&saga.Step{
		Name: "INIT_ORDER",
		Func: func(c context.Context) (err error) {
			order_dto, err = us.InitOrder(ctx, &entities.OrderEntity{
				ServiceID:    request.OrderRequest.ServiceID,
				UserID:       request.OrderRequest.UserID,
				OrderType:    request.OrderRequest.TransType,
				SubOrderType: request.OrderRequest.SubTransType,
				Amount:       request.OrderRequest.Amount,
				SourceOfFund: request.OrderRequest.SourceOfFund,
				VoucherCode:  request.OrderRequest.VoucherCode,
				DeviceID:     request.OrderRequest.DeviceID,
				BankCode:     constants.GPAY_VCCB,
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

	err = sg.AddStep(&saga.Step{
		Name: "INIT_PAYMENT",
		Func: func(c context.Context) (err error) {
			var ibft_type string
			if request.CardNo == "" {
				ibft_type = "ACCOUNT"
			} else {
				ibft_type = "CARD"
			}

			dto := &service_transaction.ETransactionDTO{
				TransactionType: constants.TRANSTYPE_WALLET_TRANS2BANK,
				ServiceType:     constants.SERVICE_TYPE_WALLET,
				TypeWallet:      constants.AmountCash,
				Amount:          request.OrderRequest.Amount,
				DeviceId:        request.OrderRequest.DeviceID,
				LastAmount:      request.OrderRequest.Amount,
				SourceOfFund:    constants.SOURCE_OF_FUND_BALANCE_WALLET,
				PayerId:         request.OrderRequest.UserID,
				GpayAccountID:   constants.GPAY_ACCOUNT_ID,
				GpayTypeWallet:  constants.Amount,
				AppId:           constants.TRANSTYPE_WALLET_TRANS2BANK,
				OrderId:         order_dto.OrderID,
				VoucherCode:     voucherId,
				BankCode:        constants.VCCB,
				AmountDiscount:  discount,
				BankTypeWallet:  constants.AmountCollectionPay,
				Message:         request.Description,
				CardNo:          request.CardNo,
				AccountNo:       request.AccountNo,
				IbftType:        ibft_type,
			}

			trans, err = us.serviceTransactionInit(ctx, dto)
			if err != nil {
				return err
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
	isTimeOutBank := false
	var bankTraceId string

	err = sg.AddStep(&saga.Step{
		Name: "IBFT TRANSFER",
		Func: func(c context.Context) (err error) {
			ibftTransferResponse, err := us.BankServiceRepository.IBFTTransfer(request.AccountNo, request.CardNo, request.OrderRequest.UserID, order_dto.OrderID,
				request.IbftCode, request.Description, request.OrderRequest.Amount, "")

			if err != nil {
				bankTraceId = ibftTransferResponse.Data.BankTraceId
				failReason = err.Error()
				return err
			}

			bankTraceId = ibftTransferResponse.Data.BankTraceId
			if ibftTransferResponse.ErrorCode.IsVerifying() {
				isTimeOutBank = true
				return errors.New("Giao dịch đang chờ xử lí")
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
			trans.BankTransactionId = bankTraceId
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
