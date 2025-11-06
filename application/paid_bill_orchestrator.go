package application

import (
	"context"
	"errors"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	errorsMap "orders-system/errors"
	"orders-system/proto/order_system"
	"orders-system/proto/service_card"
	"orders-system/proto/service_promotion"
	"orders-system/proto/service_transaction"
	"orders-system/utils/saga"

	"go.mongodb.org/mongo-driver/mongo"
)

func (us *OrderApplication) PayBill(ctx context.Context, request *order_system.OrderPayBillRequest, response *order_system.OrderPayBillResponse) (order_dto *entities.OrderEntity, err error) {
	sg := saga.NewSaga("PayBillActionAccount")

	discount := int64(0)

	voucher_id := ""
	var subTransactionType string
	var paymentTokenOrder entities.OrderEntity

	if request.OrderRequest.SourceOfFund == constants.SOURCE_OF_FUND_BANK_ATM {
		getPaymentTokenOrder, err := us.GetValidOrder(ctx, request.ConfirmPaymentTokenRequest.OrderId)
		if err != nil {
			return nil, err
		}
		paymentTokenOrder = *getPaymentTokenOrder
	}

	err = sg.AddStep(&saga.Step{
		Name: "Check Valid service code of bill",
		Func: func(c context.Context) (err error) {
			vendorInfo, err := us.ServiceCardRepository.FindVendorByCode(ctx, &service_card.FindVendorByCodeReq{
				ServiceCode: request.ServiceCode,
			})

			if err != nil {
				if err == mongo.ErrNoDocuments {
					return errors.New("Mã dịch vụ không hợp lệ")
				}
				return
			}

			switch vendorInfo.Vendor.Type {
			case "electric":
				subTransactionType = constants.SUB_TRANSTYPE_WALLET_PAY_BILL_ELECTRIC
			case "water":
				subTransactionType = constants.SUB_TRANSTYPE_WALLET_PAY_BILL_WATTER
			case "home_credit":
				subTransactionType = constants.SUB_TRANSTYPE_WALLET_PAY_BILL_LOAN
			case "tv":
				subTransactionType = constants.SUB_TRANSTYPE_WALLET_PAY_BILL_TV
			case "internet":
				subTransactionType = constants.SUB_TRANSTYPE_WALLET_PAY_BILL_INTERNET
			case "telephone":
				subTransactionType = constants.SUB_TRANSTYPE_WALLET_PAY_BILL_TELEPHONE
			}

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

	order_dto = &entities.OrderEntity{
		OrderID:             paymentTokenOrder.OrderID,
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
	}

	if paymentTokenOrder.OrderID == "" {
		err = sg.AddStep(&saga.Step{
			Name: "INIT_ORDER",
			Func: func(c context.Context) (err error) {
				order_dto, err = us.InitOrder(ctx, order_dto)
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
	}

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

	err = sg.AddStep(&saga.Step{
		Name: "ACCOUNT_VOUCHER",
		Func: func(c context.Context) (err error) {
			if request.OrderRequest.VoucherCode != "" {
				res, err := us.servicePromotionUsed(ctx, &service_promotion.UseVoucherRequest{
					Code:         order_dto.VoucherCode,
					UserId:       order_dto.UserID,
					TraceId:      order_dto.OrderID,
					Total:        1,
					Amount:       request.OrderRequest.Amount,
					ServiceCode:  subTransactionType,
					SourceOfFund: request.OrderRequest.SourceOfFund,
				})
				if err != nil {
					return err
				}
				voucher_id = res.Voucher.Voucher.Id
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

	var trans *service_transaction.ETransactionDTO
	var failReason string

	err = sg.AddStep(&saga.Step{
		Name: "INIT_PAYMENT",
		Func: func(c context.Context) (err error) {

			order_dto.SubOrderType = subTransactionType

			dto := &service_transaction.ETransactionDTO{
				TransactionType:    request.OrderRequest.TransType,
				SubTransactionType: subTransactionType,
				ServiceType:        constants.SERVICE_TYPE_WALLET,
				Amount:             request.OrderRequest.Amount,
				DeviceId:           request.OrderRequest.DeviceID,
				LastAmount:         request.OrderRequest.Amount,
				VoucherCode:        request.OrderRequest.VoucherCode,
				AmountDiscount:     discount,
				SourceOfFund:       request.OrderRequest.SourceOfFund,
				PayerId:            request.OrderRequest.UserID,
				GpayAccountID:      constants.GPAY_ACCOUNT_ID,
				GpayTypeWallet:     constants.Amount,
				AppId:              subTransactionType,
				VoucherID:          voucher_id,
				OrderId:            order_dto.OrderID,
				RefId:              request.OrderRequest.RefID,
			}

			if request.OrderRequest.SourceOfFund == constants.SOURCE_OF_FUND_BANK_ATM { // thanh toan PayByToken
				dto.BankCode = paymentTokenOrder.BankCode
			}

			if request.OrderRequest.MerchantID != "" {
				dto.SubscriberMerchantID = request.OrderRequest.MerchantID
				dto.SubscriberMerchantTypeWallet = constants.AmountCash
			}

			trans, err = us.serviceTransactionInit(ctx, dto)
			if err != nil {
				failReason = err.Error()
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

	var providerMerchantId string
	// call card service to pay bill
	isPendingSrvCard := false
	err = sg.AddStep(&saga.Step{
		Name: "PAY_BILL",
		Func: func(c context.Context) (err error) {
			payBillResponse, err := us.ServiceCardRepository.PaidBill(ctx, &service_card.PaidBillReq{
				Amount:            request.OrderRequest.Amount,
				ServiceCode:       request.ServiceCode,
				CustomerReference: request.BillingCode,
				OrderId:           order_dto.OrderID,
				UserId:            request.OrderRequest.UserID,
				SubTransType:      subTransactionType,
				AreaCode:          request.AreaCode,
				TransactionId:     trans.TransactionId,
			})

			if err != nil {
				failReason = err.Error()
				return errorsMap.ErrFailOrder
			}

			providerMerchantId = payBillResponse.ProviderMerchantId

			if payBillResponse != nil && payBillResponse.Status == constants.TRANSACTION_STATUS_PENDING {
				isPendingSrvCard = true
				return errorsMap.ErrPendingOrder
			}

			return err
		},
		CompensateFunc: func(c context.Context) (err error) {
			if isPendingSrvCard == true { // Pending Order Service Card
				if trans.Status != constants.TRANSACTION_STATUS_PENDING {
					trans, err = us.serviceTransactionPending(ctx, trans)
				}
				if !order_dto.Status.IsVerifying() {
					us.VerifyingOrder(ctx, order_dto)
				}
			} else {
				trans.FailReason = failReason
				trans, err = us.serviceTransactionCancel(ctx, trans)
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
			trans.ProviderMerchantID = providerMerchantId
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
