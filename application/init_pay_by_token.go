package application

import (
	"context"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	"orders-system/proto/order_system"
	"orders-system/proto/service_transaction"
	"orders-system/utils/saga"
)

func (us OrderApplication) InitPayByToken(ctx context.Context, req *order_system.InitPaymentTokenRequest, response *order_system.InitPaymentTokenResponse) (order entities.OrderEntity, transaction service_transaction.ETransactionDTO, err error) {
	sg := saga.NewSaga("InitPayByToken [BANK_ATM SOF]")

	var trans *service_transaction.ETransactionDTO
	var failReason, bankTraceId string

	getLinkIdInfo, err := us.BankServiceRepository.LinkInfo(req.LinkId)
	if err != nil {
		return order, transaction, err
	}

	var orderDto *entities.OrderEntity

	orderType := constants.TRANSTYPE_WALLET_PAY_BY_TOKEN
	if req.OrderType == constants.TRANSTYPE_PAY_TO_MERCHANT {
		orderType = req.OrderType
	}

	if req.OrderId != "" {
		getValidOrder, err := us.GetValidOrder(ctx, req.OrderId)
		if err != nil {
			return order, transaction, err
		}

		getValidOrder.OrderType = orderType
		getValidOrder.BankCode = getLinkIdInfo.Data.BankCode
		getValidOrder.GPayBankCode = getLinkIdInfo.Data.GpayBankCode
		getValidOrder.UserID = req.PayerId
		getValidOrder.VoucherCode = req.VoucherCode
		if req.GetSubOrderType() != "" && getValidOrder.SubOrderType != constants.SUB_TRANSTYPE_WALLET_APP_TO_APP {
			getValidOrder.SubOrderType = req.SubOrderType
		}
		getValidOrder.SourceOfFund = constants.SOURCE_OF_FUND_BANK_ATM
		orderDto = getValidOrder
	} else {
		orderDto, err = us.InitOrder(ctx, &entities.OrderEntity{
			Amount:       req.Amount,
			SourceOfFund: constants.SOURCE_OF_FUND_BANK_ATM,
			OrderType:    orderType,
			BankCode:     getLinkIdInfo.Data.BankCode,
			GPayBankCode: getLinkIdInfo.Data.GpayBankCode,
			UserID:       req.PayerId,
			SubOrderType: req.SubOrderType,
			VoucherCode:  req.VoucherCode,
		})
	}

	if err != nil {
		return order, transaction, err
	}

	bankCodeGpay := getLinkIdInfo.Data.GpayBankCode

	//todo init trans
	err = sg.AddStep(&saga.Step{
		Name: "INIT_PAYMENT",
		Func: func(c context.Context) (err error) {
			if orderDto.TransactionID != "" {
				trans, err = us.TransactionRepository.FindTransactionByID(ctx, &service_transaction.ETransactionDTO{TransactionId: orderDto.TransactionID})
				if err != nil {
					return err
				}
				transaction = *trans
			} else {
				var dto = &service_transaction.ETransactionDTO{
					TransactionType:    orderType,
					ServiceType:        constants.SERVICE_TYPE_WALLET,
					SourceOfFund:       constants.SOURCE_OF_FUND_BANK_ATM,
					PayerId:            req.PayerId,
					OrderId:            orderDto.OrderID,
					SubTransactionType: orderDto.SubOrderType,
					BankCode:           orderDto.BankCode,
					Amount:             orderDto.Amount,
					AmountDiscount:     req.DiscountAmount,
					AmountFeeGpay:      req.FeeAmount,
					VoucherCode:        req.VoucherCode,
					RefId:              orderDto.RefID,
					AppId:              req.SubOrderType,
				}
				if dto.TransactionType == constants.TRANSTYPE_PAY_TO_MERCHANT {
					dto.ProviderMerchantID = orderDto.SubscribeMerchantID
					dto.MerchantID = orderDto.SubscribeMerchantID
					dto.MerchantTypeWallet = constants.AmountRevenue
				}

				trans, err = us.serviceTransactionInit(ctx, dto)

				if err != nil {
					failReason = err.Error()
					return err
				}
				transaction = *trans
			}
			orderDto.TransactionID = transaction.TransactionId

			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			return
		},
	})
	if err != nil {
		return
	}

	var status string
	//todo payByToken request
	err = sg.AddStep(&saga.Step{
		Name: "CASH IN BANKGW",
		Func: func(c context.Context) (err error) {
			cashinRes, err := us.BankServiceRepository.CashIn(bankCodeGpay, req.LinkId, trans.LastAmount, orderDto.OrderID, req.PayerId)
			if err != nil {
				failReason = err.Error()
				return err
			}
			bankTraceId = cashinRes.Data.BankTraceID
			status = string(cashinRes.ErrorCode)
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			if trans != nil && trans.Status != constants.TRANSACTION_STATUS_FAILED && trans.Status != constants.TRANSACTION_STATUS_PENDING {
				trans.FailReason = failReason
				trans.BankTransactionId = bankTraceId
				trans, err = us.serviceTransactionCancel(ctx, trans)
			}
			if !orderDto.Status.IsVerifying() && !orderDto.Status.IsFailed() {
				orderDto.InternalErr = failReason
				_, _ = us.FailedOrder(ctx, orderDto)
			}
			return err
		},
	})
	if err != nil {
		return
	}

	//todo processing order
	err = sg.AddStep(&saga.Step{
		Name: "PROCESSING-ORDER",
		Func: func(c context.Context) (err error) {
			orderDto.BankTransactionId = bankTraceId
			orderDto, err = us.ProcessingOrder(ctx, orderDto)
			if err == nil {
				response.BankTraceId = orderDto.BankTransactionId
				response.OrderId = orderDto.OrderID
				response.Status = status
				order = *orderDto
			}
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			return err
		},
	})
	if err != nil {
		return
	}

	ordinator := saga.NewCoordinator(ctx, ctx, sg, us.LogSaga)
	rg := ordinator.Play()
	err = rg.ExecutionError
	return

}
