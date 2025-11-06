package application

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	"orders-system/domain/request_params"
	"orders-system/domain/value_objects"
	"orders-system/proto/order_system"
	"orders-system/proto/service_transaction"
	"orders-system/proto/service_user"
	"orders-system/utils/helpers"
	"orders-system/utils/telegram"
	"reflect"
)

func (us *OrderApplication) InitRefundTrans(ctx context.Context, request *order_system.RefundTransactionReq) (res *entities.OrderEntity, err error) {
	initOrder, err := us.InitOrder(ctx, &entities.OrderEntity{
		RefID:                     request.MerchantRefRefundId,
		SubscribeMerchantID:       request.MerchantId,
		ServiceType:               constants.SERVICE_TYPE_WALLET,
		OrderType:                 constants.TRANSTYPE_WALLET_REFUND,
		Amount:                    request.Amount,
		BankTransactionId:         request.RefundBankTraceId,
		Description:               "GD hoàn tiền",
		RefundSourceTransactionId: request.SourceTransactionID,
		RefundSourceOrderId:       request.SourceOrderId,
		RefundType:                request.RefundType.String(),
		MerchantCode:              request.MerchantCode,
	})

	if err != nil {
		us.Logger.With(zap.Error(err)).Error(constants.SERVICE_ORDER_SYSTEM_ERROR)
		return res, err
	}

	getTranById, err := us.TransactionRepository.FindTransactionByID(context.TODO(), &service_transaction.ETransactionDTO{
		TransactionId: request.SourceTransactionID,
	})

	// da quyen toan (settlement_status == SUCCESS) => amount_cash
	// chưa quyen toan (settlement_status != SUCCESS) => amount_revenue

	if err == nil {
		if getTranById.SourceOfFund == constants.SOURCE_OF_FUND_BALANCE_WALLET && getTranById.StatusSettlement == constants.STATUS_SUCCESS {
			if request.RefundType == order_system.RefundTransactionReq_API {
				return nil, errors.New("Hình thức hoàn tiền không hợp lệ")
			}
		}

		res, err = us.checkTransactionCondition(ctx, request, initOrder, *getTranById)
		if err != nil {
			return initOrder, err
		}
	}
	return initOrder, err
}

func (us *OrderApplication) ConfirmRefundTrans(ctx context.Context, request *order_system.ConfirmRefundTransactionReq) (getOrderByRefundId *entities.OrderEntity, err error) {
	getOrderByRefundId, err = us.OrderRepository.GetOrderByRefundId(ctx, request.RefundId)
	if err != nil {
		return getOrderByRefundId, err
	}
	bankTransactionId := request.BankTransactionId

	if request.RefundType == "API" {
		refundBankResp, err := us.BankServiceRepository.ReFund(getOrderByRefundId.RefundSourceOrderId, getOrderByRefundId.Amount)
		if err == nil {
			bankTransactionId = refundBankResp.Data.BankTransactionId
		} else {
			us.Logger.With(zap.Error(err)).Error(constants.SERVICE_BANKGW_ERROR + "_refund")
			cancelReason := constants.SERVICE_BANKGW_ERROR + "_refund" + err.Error()
			_, err = us.TransactionRepository.CancelRefundTransaction(ctx, &service_transaction.CancelRefundTransactionRequest{
				RefundId: request.RefundId,
				Reason:   cancelReason,
			})
			if err != nil {
				us.Logger.With(zap.Error(err)).Error(constants.SERVICE_TRANSACTION_ERROR + "_cancel")
			}
			return getOrderByRefundId, errors.New(cancelReason)
		}
	}

	var typeRefund service_transaction.TypeRefund

	if request.RefundType == "API" {
		typeRefund = service_transaction.TypeRefund_API
	} else if request.RefundType == "INDIRECT" {
		typeRefund = service_transaction.TypeRefund_INDIRECT
	}

	confirmRefundRes, err := us.TransactionRepository.ConfirmRefundTransaction(ctx, &service_transaction.ConfirmRefundTransactionRequest{
		RefundId:          request.RefundId,
		BankTransactionId: bankTransactionId,
		IgnoreFee:         request.IgnoreFee,
		IgnorePromotion:   request.IgnorePromotion,
		ConfirmByAdmin:    request.ConfirmByAdmin,
		TypeRefund:        typeRefund,
		RefMerchantRefund: getOrderByRefundId.RefID,
	})
	if err != nil {
		us.Logger.With(zap.Error(err)).Error(constants.SERVICE_TRANSACTION_ERROR)
		getOrderByRefundId.FailReason = err.Error()
		getOrderByRefundId, _ = us.FailedOrder(ctx, getOrderByRefundId)
		return getOrderByRefundId, err
	}
	us.Logger.Named(fmt.Sprintf("%v", confirmRefundRes)).Info("confirmRefundRes")

	err = us.sendRefundMqttNotify(ctx, confirmRefundRes.RefundTransactionId)
	if err != nil {
		return nil, err
	}
	getOrderByRefundId, _ = us.SuccessOrder(ctx, getOrderByRefundId)
	return getOrderByRefundId, err
}

func (us *OrderApplication) CancelRefundTrans(ctx context.Context, request *order_system.CancelRefundTransactionReq) (res *order_system.CancelRefundTransactionRes, err error) {
	_, err = us.TransactionRepository.CancelRefundTransaction(ctx, &service_transaction.CancelRefundTransactionRequest{
		RefundId:      request.RefundId,
		Reason:        request.CancelReason,
		CancelByAdmin: request.ConfirmByAdmin,
	})

	if err != nil {
		return res, err
	}
	return res, err
}

func (us *OrderApplication) checkTransactionCondition(ctx context.Context, request *order_system.RefundTransactionReq, order *entities.OrderEntity, trans service_transaction.ETransactionDTO) (res *entities.OrderEntity, err error) {
	var bankTransactionId string
	var typeRefund service_transaction.TypeRefund

	if request.RefundType == order_system.RefundTransactionReq_INDIRECT {
		typeRefund = service_transaction.TypeRefund_INDIRECT
	} else if request.RefundType == order_system.RefundTransactionReq_API {
		typeRefund = service_transaction.TypeRefund_API
	}

	isIgnoreFee := request.IgnoreFee
	if request.MerchantId != "" {
		isIgnoreFee = true // yc từ merchant thì mặc định có hoàn phí
	}

	initRefund, err := us.TransactionRepository.RefundTransaction(ctx, &service_transaction.RefundTransactionRequest{
		SourceTransactionId: request.SourceTransactionID,
		Amount:              request.Amount,
		Reason:              request.Reason,
		IgnoreFee:           isIgnoreFee,
		IgnorePromotion:     request.IgnorePromotion,
		Note:                request.Note,
		TypeRefund:          typeRefund,
		CreateByAdmin:       request.CreateByAdmin,
		MerchantId:          request.MerchantId,
		BankTransactionId:   request.RefundBankTraceId,
		RefMerchantRefund:   order.RefID,
	})

	if err != nil {
		us.Logger.With(zap.String("source_transaction_id", request.SourceTransactionID), zap.Error(err)).Error(constants.SERVICE_TRANSACTION_ERROR)
		order.FailReason = err.Error()
		order, _ = us.FailedOrder(ctx, order)
		return order, err
	}

	us.IPool.Submit(func() {
		if request.MerchantId != "" {
			content := telegram.SendMerchantRefund(*order, *initRefund, trans)
			telegram.SendTelegram(content, us.Config.TelegramChannelId.Refund)
		}
	})

	if order.RefundSourceOrderId == "" {
		order.RefundSourceOrderId = trans.OrderId
	}

	if request.RefundType == order_system.RefundTransactionReq_INDIRECT {
		bankTransactionId = request.RefundBankTraceId
	}

	us.Logger.Named(fmt.Sprint(initRefund)).Info("init_refund_response")

	if trans.Status == constants.TRANSACTION_STATUS_FINISH {
		//transaction merchant => check số dư tài khoản merchant
		if trans.MerchantID != "" {
			getmerchantAcc, err := us.UserRepository.GetMerchantAccount(ctx, &service_user.GetMerchantAccountReq{MerchantId: trans.MerchantID})
			if err != nil {
				us.Logger.With(zap.Any("request", trans.MerchantID)).Named(err.Error()).Error(constants.SERVICE_USER_ERROR + "_getMerchant")
				return nil, err
			}

			var notEnoughMoneyError error

			if trans.StatusSettlement != constants.STATUS_SUCCESS {
				if getmerchantAcc.AmountRevenue-request.Amount < 0 { // chưa quyen toan (settlement_status != SUCCESS) => amount_revenue
					notEnoughMoneyError = errors.New("Số dư merchant không đủ")
				}
			} else if trans.StatusSettlement == constants.STATUS_SUCCESS {
				if getmerchantAcc.AmountCash-request.Amount < 0 { // da quyen toan (settlement_status == SUCCESS) => amount_cash
					notEnoughMoneyError = errors.New("Số dư merchant không đủ")
				}
			}

			if notEnoughMoneyError != nil {
				_, err := us.TransactionRepository.CancelRefundTransaction(ctx, &service_transaction.CancelRefundTransactionRequest{
					RefundId: initRefund.RefundId,
					Reason:   notEnoughMoneyError.Error(),
				})
				if err != nil {
					us.Logger.Named(err.Error()).Error(constants.SERVICE_TRANSACTION_ERROR + "_cancelRefund")
				}

				order.FailReason = notEnoughMoneyError.Error()
				_, _ = us.FailedOrder(ctx, order)
				return nil, notEnoughMoneyError
			}
		}
		//end check số dư tài khoản merchant

		if request.MerchantId == "" && trans.StatusSettlement == constants.STATUS_SUCCESS { // yc hoan tien tren Gpay + da thanh toan thanh toan MC
			order.BankTransactionId = bankTransactionId
			order.RefundTransactionId = initRefund.RefundId
			order, err = us.ProcessingOrder(ctx, order)
			if err != nil {
				return nil, err
			}
			return order, nil
		}
		if request.MerchantId == "" && trans.StatusSettlement != constants.STATUS_SUCCESS { // yc hoan tien tren Gpay + chưa thanh toan MC
			if request.RefundType == order_system.RefundTransactionReq_API {
				refundBankResp, err := us.BankServiceRepository.ReFund(trans.OrderId, request.Amount)
				if err == nil {
					bankTransactionId = refundBankResp.Data.BankTransactionId
				} else {
					us.Logger.With(zap.Error(err)).Error(constants.SERVICE_BANKGW_ERROR)
					cancelReason := constants.SERVICE_BANKGW_ERROR + err.Error()
					_, err = us.TransactionRepository.CancelRefundTransaction(ctx, &service_transaction.CancelRefundTransactionRequest{
						RefundId: initRefund.RefundId,
						Reason:   cancelReason,
					})
					if err != nil {
						us.Logger.With(zap.Error(err)).Error(constants.SERVICE_TRANSACTION_ERROR + "_cancel")
					}
					return res, errors.New(cancelReason)
				}
			}

			var typeRefund service_transaction.TypeRefund

			if request.RefundType == order_system.RefundTransactionReq_API {
				typeRefund = service_transaction.TypeRefund_API
			} else if request.RefundType == order_system.RefundTransactionReq_INDIRECT {
				typeRefund = service_transaction.TypeRefund_INDIRECT
			}

			confirmRefundRes, err := us.TransactionRepository.ConfirmRefundTransaction(ctx, &service_transaction.ConfirmRefundTransactionRequest{
				RefundId:          initRefund.RefundId,
				BankTransactionId: bankTransactionId,
				IgnoreFee:         request.IgnoreFee,
				IgnorePromotion:   request.IgnorePromotion,
				ConfirmByAdmin:    "Hệ thống",
				TypeRefund:        typeRefund,
				RefMerchantRefund: order.RefID,
			})
			if err != nil {
				us.Logger.With(zap.Error(err)).Error(constants.SERVICE_TRANSACTION_ERROR + "_confirm")
				order.BankTransactionId = bankTransactionId
				order.FailReason = err.Error()
				order, _ = us.FailedOrder(ctx, order)
				return order, nil
			} else {
				us.Logger.Named(fmt.Sprintf("%v", confirmRefundRes)).Info("confirmRefundRes")
				order.RefundTransactionId = initRefund.RefundId
				order.BankTransactionId = bankTransactionId
				err = us.sendRefundMqttNotify(ctx, confirmRefundRes.RefundTransactionId)
				if err != nil {
					return nil, err
				}
				order, _ = us.SuccessOrder(ctx, order)
				return order, nil
			}
		} else { // check Refund config for Merchant
			order.BankTransactionId = bankTransactionId
			order.RefundTransactionId = initRefund.RefundId

			order, err = us.ProcessingOrder(ctx, order)
			if err != nil {
				return nil, err
			}

			getMerchantConfigRefundRes, err := us.IWalletConfig.GetRefundConfig(ctx, request_params.GetRefundConfigReq{
				MerchantId:   request.GetMerchantId(),
				ServiceType:  trans.ServiceType,
				TransType:    trans.TransactionType,
				SubTransType: trans.SubTransactionType,
				SourceOfFund: trans.SourceOfFund,
			})
			if err != nil {
				// khong thuc hien duyet tu dong
				return order, nil
			}

			us.Logger.Named(fmt.Sprintf("%v", getMerchantConfigRefundRes)).Info("getMerchantConfigRefundRes")

			var chooseRefundSetting value_objects.RefundSetting

			if getMerchantConfigRefundRes.TransType == trans.TransactionType {
				if len(getMerchantConfigRefundRes.SubTransType) > 0 {
					if !helpers.IsStringSliceContains(getMerchantConfigRefundRes.SubTransType, trans.SubTransactionType) {
						return order, nil
					}
				}

				for _, v := range getMerchantConfigRefundRes.Settings {
					if v.Status == "ACTIVE" && helpers.IsStringSliceContains(v.SourceOfFunds, trans.SourceOfFund) &&
						v.ConfirmType == "AUTO" {

						switch trans.StatusSettlement {
						case constants.STATUS_SUCCESS:
							if v.PaymentStatus == "PAID" {
								chooseRefundSetting = v
								break // break here
							}
						default:
							if v.PaymentStatus == "UNPAID" {
								chooseRefundSetting = v

								break // break here
							}
						}

					}
				}
			} else {
				return order, nil
			}

			us.Logger.Named(fmt.Sprintf("%v", chooseRefundSetting)).Info("chooseRefundSetting")

			g, _ := errgroup.WithContext(ctx)
			g.Go(func() error {
				if reflect.DeepEqual(chooseRefundSetting, value_objects.RefundSetting{}) {
					return errors.New("Get Empty Sastified Config ")
				}
				return nil
			})

			g.Go(func() error {
				if chooseRefundSetting.RefundCondition == "LT" {
					if request.GetAmount() >= chooseRefundSetting.RefundValue {
						return errors.New("Not sastified config ," + "request amount >= value config")
					}
				}
				return nil
			})

			g.Go(func() error {
				if chooseRefundSetting.RefundCondition == "LE" {
					if request.GetAmount() > chooseRefundSetting.RefundValue {
						return errors.New("Not sastified config ," + "request amount > value config")
					}
				}
				return nil
			})

			if err := g.Wait(); err != nil {
				us.Logger.Error(err.Error())
				return order, nil
			}

			if chooseRefundSetting.RefundType == "API" {
				refundBankResp, err := us.BankServiceRepository.ReFund(trans.OrderId, request.Amount)
				if err == nil {
					bankTransactionId = refundBankResp.Data.BankTransactionId
				} else {
					us.Logger.With(zap.Error(err)).Error(constants.SERVICE_BANKGW_ERROR)
					cancelReason := constants.SERVICE_BANKGW_ERROR + err.Error()
					_, err = us.TransactionRepository.CancelRefundTransaction(ctx, &service_transaction.CancelRefundTransactionRequest{
						RefundId: initRefund.RefundId,
						Reason:   cancelReason,
					})
					if err != nil {
						us.Logger.With(zap.Error(err)).Error(constants.SERVICE_TRANSACTION_ERROR + "_cancel")
					}
					return res, errors.New(cancelReason)
				}
			}

			var typeRefund service_transaction.TypeRefund

			if chooseRefundSetting.RefundType == "API" {
				typeRefund = service_transaction.TypeRefund_API
			} else {
				typeRefund = service_transaction.TypeRefund_INDIRECT
			}

			confirmRefundRes, err := us.TransactionRepository.ConfirmRefundTransaction(ctx, &service_transaction.ConfirmRefundTransactionRequest{
				RefundId:          initRefund.RefundId,
				BankTransactionId: bankTransactionId,
				IgnoreFee:         request.IgnoreFee,
				IgnorePromotion:   request.IgnorePromotion,
				ConfirmByAdmin:    "Hệ thống",
				TypeRefund:        typeRefund,
				RefMerchantRefund: order.RefID,
			})
			if err != nil {
				us.Logger.With(zap.Error(err)).Error(constants.SERVICE_TRANSACTION_ERROR + "_confirm")
				order.BankTransactionId = bankTransactionId
				order.FailReason = err.Error()
				order, _ = us.FailedOrder(ctx, order)
			} else {
				us.Logger.Named(fmt.Sprintf("%v", confirmRefundRes)).Info("confirmRefundRes")
				order.RefundTransactionId = initRefund.RefundId
				order.BankTransactionId = bankTransactionId
				_ = us.sendRefundMqttNotify(ctx, confirmRefundRes.RefundTransactionId)
				order, _ = us.SuccessOrder(ctx, order)
			}

			return order, err
		}
	} else {
		return nil, errors.New("Transaction's status must success")
	}

}

func (us *OrderApplication) sendRefundMqttNotify(ctx context.Context, transId string) error {
	getTranById, err := us.TransactionRepository.FindTransactionByID(context.TODO(), &service_transaction.ETransactionDTO{
		TransactionId: transId,
	})
	if err != nil {
		us.Logger.With(zap.Error(err)).Error(constants.SERVICE_TRANSACTION_ERROR + "_get_confirm_trans")
		return err
	}

	ctx = context.TODO()
	us.IPool.Submit(func() {
		us.MqttUpdateProfile(ctx, constants.UPDATE_USER_INFO, getTranById.UserReceiveRefund, getTranById.UserReceiveRefund)
		us.SendMqttTransactionByObject(ctx, constants.UPDATE_TRANSACTION, *getTranById)
		us.notificationTransaction(ctx, *getTranById)
	})
	return err
}
