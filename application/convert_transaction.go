package application

import (
	"context"
	"go.uber.org/zap"
	"net/http"
	"orders-system/domain/aggregates"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	"orders-system/domain/value_objects"
	"orders-system/proto/service_transaction"
	"orders-system/proto/service_user"
	"orders-system/utils/helpers"
	"time"
)

func (us *OrderApplication) ConvertETransToDetail(ctx context.Context, transDto service_transaction.ETransactionDTO) (res aggregates.TransactionDetail, err error) {
	var payerUser, payeeUser entities.User

	getPayer, err := us.GetProfile(ctx, transDto.PayerId)

	if err == nil {
		payerUser = getPayer.User
	}

	getPayee, err := us.GetProfile(ctx, transDto.PayeeId)

	if err == nil {
		payeeUser = getPayee.User
	}

	var old_trans_type string
	mapETransType, err := helpers.MapETransTypeToOldTrans(ctx, transDto.TransactionType, transDto.SubTransactionType)
	if err == nil || mapETransType == "" {
		old_trans_type = mapETransType
	}

	var merchantCode string
	if transDto.MerchantID != "" {
		getMerchant, err := us.UserRepository.FindMerchantAccountByID(helpers.ContextWithTimeOut(), &service_user.FindMerchantAccountByIdRequest{Id: transDto.MerchantID})
		if err == nil {
			merchantCode = getMerchant.MerchantDetail.Merchant.Code
		} else {
			us.Logger.With(zap.Any("request", transDto.MerchantID)).With(zap.Any("error", err.Error())).Error("err_get_merchant")
		}
	}
	var redirectUrl string
	if helpers.IsStringSliceContains([]string{constants.SUB_TRANSTYPE_WALLET_WEB_IN_APP,
		constants.SUB_TRANSTYPE_WALLET_APP_TO_APP,
		constants.SUB_TRANSTYPE_WALLET_WEB_TO_APP}, transDto.SubTransactionType) {
		var responseRedirectUrl value_objects.RedirectUrl
		mainPathUrl := "order-qr/get-callback/"
		if transDto.SubTransactionType == constants.SUB_TRANSTYPE_WALLET_WEB_IN_APP {
			mainPathUrl = "/order/web-in-app/get-callback/"
		}

		err = helpers.HttpRequest(struct {
			Uri      string
			Path     string
			Method   string
			Headers  map[string]string
			Body     interface{}
			Response interface{}
		}{
			Uri:      us.Config.MerchantCallBack.Uri,
			Path:     mainPathUrl + transDto.OrderId,
			Method:   http.MethodGet,
			Headers:  map[string]string{us.Config.MerchantCallBack.SecretKey: us.Config.MerchantCallBack.SecretValue},
			Response: &responseRedirectUrl,
		})
		if err == nil {
			redirectUrl = responseRedirectUrl.Response.RedirectUrl
		} else {
			us.Logger.Error("err_redirect_app", zap.Error(err))
		}
	}

	if transDto.TransactionType == constants.TRANSTYPE_PAY_VA && transDto.Status == constants.TRANSACTION_STATUS_FINISH {
		go func() {
			var r interface{}
			err = helpers.HttpRequest(struct {
				Uri      string
				Path     string
				Method   string
				Headers  map[string]string
				Body     interface{}
				Response interface{}
			}{
				Uri:      us.Config.MerchantCallBack.Uri,
				Path:     "virtual-account/resend-va-transaction/" + transDto.OrderId,
				Method:   http.MethodGet,
				Headers:  map[string]string{us.Config.MerchantCallBack.SecretKey: us.Config.MerchantCallBack.SecretValue},
				Response: r,
			})
		}()
	}

	return aggregates.TransactionDetail{
		Transaction: entities.Transaction{
			Id:                        transDto.TransactionId,
			TransactionId:             transDto.TransactionId,
			AppId:                     transDto.AppId,
			Currency:                  transDto.Currency,
			Message:                   transDto.Message,
			State:                     transDto.State,
			Status:                    transDto.Status,
			ServiceType:               transDto.ServiceType,
			TransactionType:           transDto.TransactionType,
			SubTransactionType:        transDto.SubTransactionType,
			TypeWallet:                transDto.TypeWallet,
			Amount:                    uint64(transDto.Amount),
			LastAmount:                transDto.LastAmount,
			FeeGpay:                   transDto.AmountFeeGpay,
			DiscountAmount:            transDto.AmountDiscount,
			AmountMerchantFee:         transDto.AmountMerchantFee,
			AmountTransactionCashBack: transDto.AmountTransactionCashBack,
			VoucherCode:               transDto.VoucherCode,
			Source:                    transDto.SourceOfFund,
			MerchantId:                transDto.MerchantID,
			MerchantTransactionId:     transDto.MerchantTransactionId,
			TransactionCashback:       transDto.CashbackTransactionID,
			RefId:                     transDto.RefId,
			OrderId:                   transDto.OrderId,
			DeviceId:                  transDto.DeviceId,
			PayerId:                   transDto.PayerId,
			PayeeId:                   transDto.PayeeId,
			PayerStatusKyc:            transDto.PayerStatusKyc,
			PayerStatusLinkedBank:     transDto.PayerStatusLinkedBank,
			PayeeStatusKyc:            transDto.PayeeStatusKyc,
			PayeeStatusLinkedBank:     transDto.PayeeStatusLinkedBank,
			GpayAccountID:             transDto.GpayAccountID,
			BankTransactionId:         transDto.BankTransactionId,
			Napas:                     transDto.Napas,
			BankCode:                  transDto.BankCode,
			IbftType:                  transDto.IbftType,
			CardNo:                    transDto.CardNo,
			AccountNo:                 transDto.AccountNo,
			FailReason:                transDto.FailReason,
			ReturnCode:                transDto.ReturnCode,
			ReturnMessage:             transDto.ReturnMessage,
			Exception:                 transDto.Exception,
			CreatedAt:                 time.Unix(transDto.CreatedAt, 0),
			UpdatedAt:                 time.Unix(transDto.UpdatedAt, 0),
			DeletedAt:                 time.Unix(transDto.DeletedAt, 0),

			Type:                   old_trans_type,
			AmountAfterAddVoucher:  0,
			AmountBeforeAddVoucher: 0,
			InvoiceId:              "",
			TypeTransfer:           "",
			PayerAmountBefore:      0,
			PayeeAmountBefore:      0,
			PayerAmountAfter:       0,
			PayeeAmountAfter:       0,
			UnUseWallet:            false,
			PartnerCode:            "",
			WalletIdSrc:            "",
			WalletIdDist:           "",
			V:                      0,
			PaymentType:            "",
			ProviderMerchantId:     transDto.ProviderMerchantID,
			MerchantCode:           merchantCode,
			RedirectUrl:            redirectUrl,
		},
		Payer: payerUser,
		Payee: payeeUser,
	}, err
}
