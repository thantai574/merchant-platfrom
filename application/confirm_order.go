package application

import (
	"context"
	"errors"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	"orders-system/proto/order_system"
	"orders-system/proto/service_promotion"
	"orders-system/proto/service_transaction"
	mapErrString "orders-system/utils/errors"
	"orders-system/utils/helpers"
	"orders-system/utils/saga"
)

func (us *OrderApplication) ConfirmOrder(ctx context.Context, request *order_system.ConfirmOrderRequest, response *order_system.ConfirmOrderResponse) (order_dto *entities.OrderEntity, err error) {
	sg := saga.NewSaga("Confirm Order")

	getOrderById, err := us.GetValidOrder(ctx, request.OrderId)
	if err != nil {
		return order_dto, err
	}

	order_dto = getOrderById

	if request.OrderRequest != nil {
		if request.GetOrderRequest().GetSubTransType() != "" && order_dto.SubOrderType != constants.SUB_TRANSTYPE_WALLET_APP_TO_APP {
			order_dto.ServiceCode = request.GetOrderRequest().GetSubTransType()
		}

		if request.OrderRequest.VoucherCode != "" {
			order_dto.VoucherCode = request.OrderRequest.VoucherCode
		}
		if request.OrderRequest.UserID != "" {
			order_dto.UserID = request.OrderRequest.UserID
		}
		order_dto.DeviceID = request.OrderRequest.DeviceID
	}

	err = sg.AddStep(&saga.Step{
		Name: "PROCESSING_ORDER",
		Func: func(c context.Context) (err error) {
			if order_dto.OrderType == constants.TRANSTYPE_OPENWAL_TRANSFERTOBANK {
				order_dto.BankReceived = order_dto.BankCode
				order_dto.BankCode = constants.VCCB
			}

			order_dto, err = us.ProcessingOrder(ctx, order_dto)
			return
		},
		CompensateFunc: func(c context.Context) (err error) {
			if !order_dto.Status.IsVerifying() && !order_dto.Status.IsFailed() {
				_, _ = us.FailedOrder(ctx, order_dto)
			}
			return
		},
		Options: nil,
	})
	if err != nil {
		return
	}

	var voucherId string
	var discountUserAmount int64
	var failReason, bankTraceId string
	var isTimeOutBank bool
	var trans *service_transaction.ETransactionDTO

	//@todo Voucher
	err = sg.AddStep(&saga.Step{
		Name: "ACCOUNT_VOUCHER",
		Func: func(c context.Context) (err error) {
			if getOrderById.VoucherCode != "" {
				res, err := us.servicePromotionUsed(ctx, &service_promotion.UseVoucherRequest{
					Code:         order_dto.VoucherCode,
					UserId:       order_dto.UserID,
					TraceId:      order_dto.OrderID,
					Total:        1,
					Amount:       order_dto.Amount,
					ServiceCode:  order_dto.SubOrderType,
					SourceOfFund: order_dto.SourceOfFund,
				})
				if err != nil {
					return err
				}
				voucherId = res.Voucher.Voucher.Id
				discountUserAmount = res.DiscountAmount
			}
			return err
		},
		CompensateFunc: func(c context.Context) (err error) {
			_, err = us.servicePromotionCompensate(ctx, &service_promotion.ReverseWalletRequest{
				TraceId: order_dto.OrderID,
			})
			return
		},
		Options: nil,
	})
	if err != nil {
		return
	}

	//@todo Trans
	err = sg.AddStep(&saga.Step{
		Name: "INIT_PAYMENT",
		Func: func(c context.Context) (err error) {
			dto := &service_transaction.ETransactionDTO{
				AppId:                    order_dto.ServiceCode,
				ServiceType:              order_dto.ServiceType,
				TransactionType:          order_dto.OrderType,
				SubTransactionType:       order_dto.SubOrderType,
				Amount:                   order_dto.Amount,
				LastAmount:               order_dto.Amount,
				AmountDiscount:           discountUserAmount,
				VoucherCode:              order_dto.VoucherCode,
				SourceOfFund:             order_dto.SourceOfFund,
				RefId:                    order_dto.RefID,
				OrderId:                  order_dto.OrderID,
				DeviceId:                 order_dto.DeviceID,
				PayerId:                  order_dto.UserID,
				PayeeId:                  order_dto.ToUserID,
				BankCode:                 order_dto.BankCode,
				VoucherID:                voucherId,
				Napas:                    order_dto.Napas,
				BankTransactionId:        order_dto.BankTransactionId,
				MerchantID:               order_dto.SubscribeMerchantID,
				AmountMerchantFee:        order_dto.AmountMerchantFee,
				AmountMerchantFeeGpayTmp: order_dto.AmountMerchantFeeGpayTmp,
				MerchantTypeWallet:       order_dto.MerchantTypeWallet,
				FixedFeeAmount:           order_dto.FixedFeeAmount,
				RateFeeAmount:            order_dto.RateFeeAmount,
				MerchantFeeMethod:        order_dto.MerchantFeeMethod,
				IbftReceiveBank:          order_dto.BankReceived,
				AccountNo:                order_dto.AccountNo,
				CardNo:                   order_dto.CardNumber,
			}

			if order_dto.ServiceType == "" {
				dto.ServiceType = constants.SERVICE_TYPE_WALLET
			}

			switch order_dto.OrderType {
			case constants.TRANSTYPE_PAY_TO_MERCHANT:
				dto.ProviderMerchantID = order_dto.SubscribeMerchantID
				dto.MerchantTypeWallet = constants.AmountRevenue
			case constants.TRANSTYPE_PG_PAY:
				dto.SubscriberMerchantID = ""
			case constants.TRANSTYPE_PG_PAY_BY_PAY_INTERNATIONAL_CARD:
				dto.MerchantTypeWallet = constants.AmountRevenue
			default:

			}

			if request.OrderRequest != nil {
				if request.OrderRequest.SubTransType != "" && order_dto.SubOrderType != constants.SUB_TRANSTYPE_WALLET_APP_TO_APP {
					order_dto.SubOrderType = request.OrderRequest.SubTransType
					dto.SubTransactionType = request.OrderRequest.SubTransType
				}
			}

			trans, err = us.serviceTransactionInit(ctx, dto)
			if err != nil {
				failReason = mapErrString.GetGrpcErrMessage(err)
				return errors.New(failReason)
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

	// todo case OPEN WALLET
	err = sg.AddStep(&saga.Step{
		Name: "OPEN WALLET",
		Func: func(c context.Context) (err error) {
			if helpers.IsStringSliceContains([]string{constants.TRANSTYPE_OPENWAL_TRANSFERTOBANK}, order_dto.OrderType) {
				ibftTransferResponse, err := us.BankServiceRepository.IBFTTransfer(order_dto.AccountNo, order_dto.CardNumber, order_dto.SubscribeMerchantID,
					order_dto.OrderID, "", order_dto.Description, order_dto.Amount, order_dto.BankCode)

				if err != nil {
					bankTraceId = ibftTransferResponse.Data.BankTraceId
					failReason = err.Error()
					return err
				}

				bankTraceId = ibftTransferResponse.Data.BankTraceId
				if ibftTransferResponse.ErrorCode.IsVerifying() {
					isTimeOutBank = true
					err := errors.New("Giao dịch đang chờ xử lí")
					order_dto.InternalErr = constants.SERVICE_BANKGW_ERROR + err.Error()
					return err
				}
				return err
			}
			return err
		},
		CompensateFunc: func(c context.Context) (err error) {
			if isTimeOutBank == true {
				if trans.Status != constants.TRANSACTION_STATUS_PENDING {
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
		return nil, err
	}

	//todo Confirm trans
	err = sg.AddStep(&saga.Step{
		Name: "FINALIZATION_PAYMENT",
		Func: func(c context.Context) (err error) {
			order_dto.TransactionID = trans.TransactionId
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
		return
	}

	//todo Confirm order
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
