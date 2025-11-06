package application

import (
	"context"
	builtinerr "errors"
	"fmt"
	"go.uber.org/zap"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	"orders-system/proto/order_system"
	"orders-system/proto/service_promotion"
	"orders-system/proto/service_transaction"
	"orders-system/utils/saga"
)

func (us *OrderApplication) ConfirmOrderWithToken(ctx context.Context, request *order_system.ConfirmOrderRequest, response *order_system.ConfirmOrderResponse) (orderDto *entities.OrderEntity, err error) {
	sg := saga.NewSaga("QRWithToken With Token ")

	var trans *service_transaction.ETransactionDTO
	var isFailedOrder, isTimeOutBank bool
	var failReason, bankTraceId string
	var voucherId string

	orderDto = new(entities.OrderEntity)

	if request.ConfirmPaymentTokenRequest.Otp != "" {
		//todo Get order
		err = sg.AddStep(&saga.Step{
			Name: "FIND_ORDER_BY_ID",
			Func: func(c context.Context) (err error) {
				findOrder, err := us.GetValidOrder(ctx, request.ConfirmPaymentTokenRequest.OrderId)
				if err != nil {
					return err
				}
				orderDto = findOrder
				return
			},
			CompensateFunc: func(c context.Context) (err error) {
				return err
			},
		})
		if err != nil {
			return
		}

		//todo Get transaction
		err = sg.AddStep(&saga.Step{
			Name: "GET INIT TRANSACTION",
			Func: func(c context.Context) (err error) {

				findTranById, err := us.TransactionRepository.FindTransactionByID(ctx, &service_transaction.ETransactionDTO{TransactionId: orderDto.TransactionID})
				if err != nil {
					return err
				}

				findTranById.MerchantID = orderDto.SubscribeMerchantID
				findTranById.BankTransactionId = orderDto.BankTransactionId
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

		err = sg.AddStep(&saga.Step{
			Name: "ACCOUNT_VOUCHER",
			Func: func(c context.Context) (err error) {
				if request.OrderRequest.VoucherCode != "" {
					res, err := us.servicePromotionUsed(ctx, &service_promotion.UseVoucherRequest{
						Code:         orderDto.VoucherCode,
						UserId:       orderDto.UserID,
						TraceId:      orderDto.OrderID,
						Total:        1,
						Amount:       orderDto.Amount,
						ServiceCode:  orderDto.SubOrderType,
						SourceOfFund: orderDto.SourceOfFund,
					})
					if err != nil {
						return err
					}
					voucherId = res.Voucher.Voucher.Id
				}
				return err
			},
			CompensateFunc: func(c context.Context) (err error) {
				_, err = us.servicePromotionCompensate(ctx, &service_promotion.ReverseWalletRequest{
					TraceId: orderDto.OrderID,
				})
				return
			},
			Options: nil,
		})
		if err != nil {
			return
		}

		//todo Verify OTP
		err = sg.AddStep(&saga.Step{
			Name: "VERITY OTP",
			Func: func(c context.Context) (err error) {
				if request.ConfirmPaymentTokenRequest.Otp != "" {
					cashInVerifyOTP, err := us.ConfirmPayByToken(ctx, order_system.ConfirmPaymentTokenRequest{
						BankTraceId:  request.ConfirmPaymentTokenRequest.BankTraceId,
						OrderId:      request.ConfirmPaymentTokenRequest.OrderId,
						LinkId:       request.ConfirmPaymentTokenRequest.LinkId,
						Otp:          request.ConfirmPaymentTokenRequest.Otp,
						GpayBankCode: orderDto.GPayBankCode,
					})

					if err != nil {
						isFailedOrder = true
						bankTraceId = cashInVerifyOTP.Data.BankTraceID
						failReason = err.Error()
						return err
					}

					bankTraceId = cashInVerifyOTP.Data.BankTraceID
					if cashInVerifyOTP.ErrorCode.IsVerifying() {
						isTimeOutBank = true
						return builtinerr.New(cashInVerifyOTP.Message)
					}

					if cashInVerifyOTP.ErrorCode.IsWrongOTP() {
						isFailedOrder = false
						return builtinerr.New(cashInVerifyOTP.Message)
					}
				}
				return err
			},
			CompensateFunc: func(c context.Context) (err error) {
				trans.BankTransactionId = bankTraceId
				if isTimeOutBank == true {
					if !orderDto.Status.IsVerifying() {
						us.VerifyingOrder(ctx, orderDto)
					}
					if trans.Status != constants.TRANSACTION_STATUS_PENDING {
						trans, err = us.serviceTransactionPending(ctx, trans)
					}
				} else {
					if isFailedOrder == true {
						trans.FailReason = failReason
						if trans != nil && trans.Status != constants.TRANSACTION_STATUS_PENDING && trans.Status != constants.TRANSACTION_STATUS_FINISH &&
							trans.Status != constants.TRANSACTION_STATUS_FAILED {
							trans, err = us.serviceTransactionCancel(ctx, trans)

							orderDto.InternalErr = failReason
							us.FailedOrder(ctx, orderDto)
						}
					}
				}
				return
			},
		})
		if err != nil {
			return
		}
	} else {
		ress := order_system.InitPaymentTokenResponse{}
		var discountAmount int64

		findOrder, err := us.GetValidOrder(ctx, request.ConfirmPaymentTokenRequest.OrderId)
		if err != nil {
			return nil, err
		}

		if request.OrderRequest.VoucherCode != "" {
			checkFeeQuota, err := us.TransactionRepository.CheckTransactionQuotaAndFee(ctx, &service_transaction.CheckTransactionQuotaAndFeeReq{
				Amount:       request.OrderRequest.Amount,
				UserId:       request.OrderRequest.UserID,
				ServiceType:  constants.SERVICE_TYPE_WALLET,
				TransType:    constants.TRANSTYPE_WALLET_PAY_BY_TOKEN,
				SubTransType: request.OrderRequest.SubTransType,
				SourceOfFund: constants.SOURCE_OF_FUND_BANK_ATM,
				IsVerify:     true,
				IsLinkedBank: true,
				VoucherCode:  request.OrderRequest.VoucherCode,
				ServiceCode:  request.OrderRequest.SubTransType,
			})
			if err != nil {
				us.Logger.With(zap.Error(err)).Error(constants.SERVICE_TRANSACTION_ERROR + "_check_quota_fee")
				return orderDto, err
			}
			discountAmount = checkFeeQuota.DiscountAmount
		}

		order, transaction, err := us.InitPayByToken(ctx, &order_system.InitPaymentTokenRequest{
			LinkId:         request.ConfirmPaymentTokenRequest.LinkId,
			Amount:         request.OrderRequest.Amount,
			PayerId:        request.OrderRequest.UserID,
			SubOrderType:   request.OrderRequest.SubTransType,
			VoucherCode:    request.OrderRequest.VoucherCode,
			OrderType:      findOrder.OrderType,
			DiscountAmount: discountAmount,
			OrderId:        findOrder.OrderID,
		}, &ress)

		if err != nil {
			us.Logger.With(zap.Error(err)).Error(constants.SERVICE_BANKGW_ERROR + "_cashin")
			return nil, err
		}

		trans = &transaction
		orderDto = &order
		bankTraceId = ress.BankTraceId

		//todo Fail Order if status cash in != "200"
		err = sg.AddStep(&saga.Step{
			Name: "FAIL ORDER",
			Func: func(c context.Context) (err error) {
				if ress.Status != "200" {
					trans.BankTransactionId = bankTraceId
					trans.FailReason = constants.SERVICE_BANKGW_ERROR + "status " + ress.Status
					_, err := us.serviceTransactionCancel(ctx, &transaction)
					if err != nil {
						return err
					}

					_, err = us.FailedOrder(ctx, &order)
					return fmt.Errorf("%v", "Giao dịch thất bại")
				}
				return err
			},
			CompensateFunc: func(c context.Context) (err error) {
				return
			},
			Options: nil,
		})
		if err != nil {
			return nil, err
		}

		//todo voucher
		err = sg.AddStep(&saga.Step{
			Name: "ACCOUNT_VOUCHER",
			Func: func(c context.Context) (err error) {
				if request.OrderRequest.VoucherCode != "" {
					res, err := us.servicePromotionUsed(ctx, &service_promotion.UseVoucherRequest{
						Code:         order.VoucherCode,
						UserId:       order.UserID,
						TraceId:      order.OrderID,
						Total:        1,
						Amount:       request.OrderRequest.Amount,
						ServiceCode:  request.OrderRequest.SubTransType,
						SourceOfFund: request.OrderRequest.SourceOfFund,
					})
					if err != nil {
						return err
					}
					voucherId = res.Voucher.Voucher.Id
				}
				return err
			},
			CompensateFunc: func(c context.Context) (err error) {
				_, err = us.servicePromotionCompensate(ctx, &service_promotion.ReverseWalletRequest{
					TraceId: order.OrderID,
				})
				return
			},
			Options: nil,
		})
		if err != nil {
			return nil, err
		}
	}

	//todo Confirm transaction
	err = sg.AddStep(&saga.Step{
		Name: "FINALIZATION_PAYMENT",
		Func: func(c context.Context) (err error) {
			trans.BankTransactionId = bankTraceId
			trans.VoucherID = voucherId
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

	//todo success order
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
			if trans.Status != constants.TRANSACTION_STATUS_PENDING {
				trans, err = us.serviceTransactionPending(ctx, trans)
			}
			if !orderDto.Status.IsVerifying() {
				us.VerifyingOrder(ctx, orderDto)
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
