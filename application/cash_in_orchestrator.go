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

func (us *OrderApplication) CashIn(ctx context.Context, request *order_system.OrderCashInRequest, response *order_system.OrderCashInResponse) (order_dto *entities.OrderEntity, err error) {
	sg := saga.NewSaga("CashInRequest")

	order_dto = new(entities.OrderEntity)
	var bankCodeGpay, bankCode string
	var isByPassOtp bool

	err = sg.AddStep(&saga.Step{
		Name: "Check VALID LinkID and User ID",
		Func: func(c context.Context) (err error) {
			getLinkIdInfo, err := us.BankServiceRepository.LinkInfo(request.LinkId)
			if err != nil {
				return err
			}

			if getLinkIdInfo.Data.GpayUserID != request.OrderRequest.ToUserID {
				return errors.New("Invalid link id ")
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

	if err != nil {
		return
	}

	err = sg.AddStep(&saga.Step{
		Name: "INIT_ORDER",
		Func: func(c context.Context) (err error) {
			orderType := request.OrderRequest.TransType
			if orderType == "" {
				orderType = constants.TRANSTYPE_WALLET_CASH_IN
			}

			serviceType := request.OrderRequest.ServiceType
			if serviceType == "" {
				serviceType = constants.SERVICE_TYPE_WALLET
			}

			order_dto, err = us.InitOrder(ctx, &entities.OrderEntity{
				ServiceID:           request.OrderRequest.ServiceID,
				UserID:              request.OrderRequest.UserID,
				SubscribeMerchantID: request.OrderRequest.MerchantID,
				ServiceType:         serviceType,
				OrderType:           orderType,
				SubOrderType:        request.OrderRequest.SubTransType,
				Amount:              request.OrderRequest.Amount,
				Quantity:            request.OrderRequest.Quantity,
				SourceOfFund:        constants.SOURCE_OF_FUND_BANK_ATM,
				VoucherCode:         request.OrderRequest.VoucherCode,
				DeviceID:            request.OrderRequest.DeviceID,
				BankCode:            bankCode,
				ToUserID:            request.OrderRequest.ToUserID,
				ServiceCode:         constants.TRANSTYPE_WALLET_CASH_IN,
				RefID:               request.OrderRequest.RefID,
			})

			if err == nil {
				response.OrderEntity = order_dto.ConvertToProto()
			}

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
				TransactionType: order_dto.OrderType,
				ServiceType:     order_dto.ServiceType,
				Amount:          order_dto.Amount,
				SourceOfFund:    order_dto.SourceOfFund,
				PayeeId:         order_dto.ToUserID,
				AppId:           order_dto.ServiceCode,
				OrderId:         order_dto.OrderID,
				BankCode:        order_dto.BankCode,
				RefId:           order_dto.RefID,
				MerchantID:      order_dto.SubscribeMerchantID,
			}

			if dto.ServiceType == "" {
				dto.ServiceType = constants.SERVICE_TYPE_WALLET
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
				us.FailedOrder(ctx, order_dto)
			}
			return
		},
	})

	if err != nil {
		return
	}

	err = sg.AddStep(&saga.Step{
		Name: "PROCESSING_ORDER",
		Func: func(c context.Context) (err error) {
			order_dto, err = us.ProcessingOrder(ctx, order_dto)
			if err == nil {
				response.OrderEntity = order_dto.ConvertToProto()
			}
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			if !order_dto.Status.IsVerifying() && !order_dto.Status.IsFailed() {
				us.CancelOrder(ctx, order_dto)
			}
			return
		},
		Options: nil,
	})

	if err != nil {
		return
	}

	isTimeOutBank := false
	var bankTraceId string

	err = sg.AddStep(&saga.Step{
		Name: "CASH IN REQUEST",
		Func: func(c context.Context) (err error) {
			cashinRes, err := us.BankServiceRepository.CashIn(bankCodeGpay, request.LinkId, request.OrderRequest.Amount, order_dto.OrderID,
				request.OrderRequest.ToUserID)
			if err != nil {
				failReason = err.Error()
				return err
			}

			bankTraceId = cashinRes.Data.BankTraceID
			response.BankTraceId = bankTraceId
			response.LinkId = request.LinkId
			order_dto.BankTransactionId = bankTraceId
			if bankTraceId == "" {
				order_dto.BankTransactionId = order_dto.OrderID
			}

			if cashinRes.ErrorCode.IsNeedToEnterOTP() {
				response.Status = order_system.OrderCashInResponse_NeedToVerifyOtp
			} else if cashinRes.ErrorCode.IsSuccess() {
				response.Status = order_system.OrderCashInResponse_Success
				isByPassOtp = true
			}
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			if isTimeOutBank == true {
				if !order_dto.Status.IsVerifying() {
					us.VerifyingOrder(ctx, order_dto)
				}
			} else {
				_, err = us.FailedOrder(ctx, order_dto)
			}
			return
		},
	})
	if err != nil {
		return
	}

	err = sg.AddStep(&saga.Step{
		Name: "CONFIRM TRANS IF BY PASS OTP",
		Func: func(c context.Context) (err error) {
			if isByPassOtp {
				trans.BankTransactionId = order_dto.BankTransactionId
				trans, err = us.serviceTransactionConfirm(ctx, trans)
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
	})
	if err != nil {
		return
	}

	//todo Confirm Order
	err = sg.AddStep(&saga.Step{
		Name: "SUCCESS_ORDER",
		Func: func(c context.Context) (err error) {
			if isByPassOtp {
				order_dto, err = us.SuccessOrder(ctx, order_dto)
				if err == nil {
					response.OrderEntity = order_dto.ConvertToProto()
				}
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
