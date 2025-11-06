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

func (us *OrderApplication) FundWal2Bank(ctx context.Context, request *order_system.FundWallet2BankRequest, response *order_system.FundWallet2BankResponse) (order_dto *entities.OrderEntity, err error) {
	sg := saga.NewSaga("Fund WAL2BANK")

	var discount int64 = 0
	var bankTraceId string
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
				PhoneTopUp:          request.OrderRequest.PhoneTopup,
				BankCode:            request.BankCode,
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

	var failReason string
	var trans *service_transaction.ETransactionDTO

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
				TransactionType: constants.TRANSTYPE_BEHALF_PAY_WALLET_TO_BANK,
				ServiceType:     constants.SERVICE_TYPE_COLLECTION_AND_PAY,
				Amount:          request.OrderRequest.Amount,
				DeviceId:        request.OrderRequest.DeviceID,
				LastAmount:      request.OrderRequest.Amount,
				VoucherCode:     request.OrderRequest.VoucherCode,
				SourceOfFund:    constants.SOURCE_OF_FUND_BALANCE_WALLET,
				GpayAccountID:   constants.GPAY_ACCOUNT_ID,
				GpayTypeWallet:  constants.Amount,
				AppId:           constants.TRANSTYPE_BEHALF_PAY_WALLET_TO_BANK,
				VoucherID:       voucherId,
				AmountDiscount:  discount,
				OrderId:         order_dto.OrderID,
				RefId:           request.OrderRequest.RefID,
				Message:         request.Description,
				CardNo:          request.CardNo,
				AccountNo:       request.AccountNo,
				BankCode:        constants.VCCB,
				BankTypeWallet:  constants.AmountCollectionPay,
				IbftReceiveBank: request.BankCode,
				IbftType:        ibft_type,
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
			return err
		},
		CompensateFunc: func(c context.Context) (err error) {
			if trans != nil && trans.Status != constants.TRANSACTION_STATUS_FAILED && trans.Status != constants.TRANSACTION_STATUS_PENDING {
				trans.BankTransactionId = bankTraceId
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
	isTimeOutBank := false

	// call IBFT Bank service
	err = sg.AddStep(&saga.Step{
		Name: "IBFT TRANSFER",
		Func: func(c context.Context) (err error) {
			ibftTransferResponse, err := us.BankServiceRepository.IBFTTransfer(request.AccountNo, request.CardNo, request.OrderRequest.MerchantID,
				order_dto.OrderID, request.IbftCode, request.Description, request.OrderRequest.Amount, request.BankName)

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
					_, _ = us.VerifyingOrder(ctx, order_dto)
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
				_, _ = us.VerifyingOrder(ctx, order_dto)
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
