package application

import (
	"context"
	"errors"
	"fmt"
	"orders-system/domain/constants"
	"orders-system/proto/service_merchant_fee"
	"orders-system/proto/service_transaction"
	"orders-system/proto/service_user"
	"orders-system/utils/helpers"

	"github.com/dustin/go-humanize"
	"github.com/leekchan/accounting"
	"go.uber.org/zap"
)

func (us *OrderApplication) serviceTransactionInit(ctx context.Context, request *service_transaction.ETransactionDTO) (res *service_transaction.ETransactionDTO, err error) {
	if request.PayerId != "" {
		u, e := us.GetProfile(ctx, request.PayerId)
		if e == nil {
			links, err := us.BankServiceRepository.LinkList(request.PayerId)
			if err != nil {
				us.Logger.Error("err_get_link_list", zap.String("err_get_link_list", err.Error()))
			}
			if u.KYC != "ACTIVE" {
				return res, errors.New("Chưa xác thực KYC")
			}

			if !request.Napas {
				if !u.HasEverLinkedBank {
					return res, errors.New("Chưa từng liên kết ngân hàng ")
				}
			}

			request.PayerStatusLinkedBank = len(links.Data) > 0
			request.PayerStatusKyc = u.KYC == "ACTIVE"
		}
	}

	if request.PayeeId != "" {
		u, e := us.GetProfile(ctx, request.PayeeId)
		if e == nil {
			links, err := us.BankServiceRepository.LinkList(request.PayeeId)
			if err != nil {
				us.Logger.Error("err_get_link_list", zap.String("err_get_link_list", err.Error()))
			}
			request.PayeeStatusLinkedBank = len(links.Data) > 0

			//if request.TransactionType == constants.TRANSTYPE_WALLET_LIXI {

			if u.KYC != "ACTIVE" {
				return res, errors.New("Người nhận chưa xác thực KYC")
			}
			if !request.Napas {
				if !u.HasEverLinkedBank {
					return res, errors.New("Người nhận chưa từng liên kết ngân hàng ")
				}
			}

			request.PayeeStatusKyc = u.KYC == "ACTIVE"
		}
	}

	checkUserID := request.PayerId
	isVerify := request.PayerStatusKyc
	isLinkedBank := request.PayerStatusLinkedBank

	if helpers.IsStringSliceContains([]string{
		constants.TRANSTYPE_WALLET_CASH_IN,
		constants.TRANSTYPE_OPENWAL_CASHIN,
		constants.TRANSTYPE_WALLET_LIXI}, request.TransactionType) {
		checkUserID = request.PayeeId
		isVerify = request.PayeeStatusKyc
		isLinkedBank = request.PayeeStatusLinkedBank
	}

	userFee, err := us.TransactionRepository.CheckTransactionQuotaAndFee(ctx, &service_transaction.CheckTransactionQuotaAndFeeReq{
		Amount:       request.Amount,
		MerchantId:   request.MerchantID,
		UserId:       checkUserID,
		ServiceType:  request.ServiceType,
		TransType:    request.TransactionType,
		SubTransType: request.SubTransactionType,
		SourceOfFund: request.SourceOfFund,
		IsVerify:     isVerify,
		IsLinkedBank: isLinkedBank,
		ServiceCode:  request.AppId,
	})

	if err != nil {
		us.Logger.Error(err.Error())
		return res, err
	}
	// check merchant quota fee
	// stop check fee if transaction already has fee
	if request.MerchantID != "" &&
		request.AmountMerchantFee == 0 && request.AmountMerchantFeeGpayTmp == 0 {
		var vaAccountType service_merchant_fee.CheckMerchantQuotaAndFeeReq_VAType

		if request.TransactionType == constants.TRANSTYPE_PAY_VA {
			switch request.AccountVAType {
			case constants.VA_ACCOUNT_TYPE_ONCE:
				vaAccountType = service_merchant_fee.CheckMerchantQuotaAndFeeReq_ONETIME
			case constants.VA_ACCOUNT_TYPE_MANY:
				vaAccountType = service_merchant_fee.CheckMerchantQuotaAndFeeReq_MANYTIME
			default:
				vaAccountType = service_merchant_fee.CheckMerchantQuotaAndFeeReq_OPTIONAL
			}
		}

		merchantFee, err := us.MerchantFeeRepository.CheckMerchantQuotaAndFee(ctx, &service_merchant_fee.CheckMerchantQuotaAndFeeReq{
			Amount:       request.Amount,
			ServiceType:  request.ServiceType,
			TransType:    request.TransactionType,
			SubTransType: request.SubTransactionType,
			MerchantId:   request.MerchantID,
			SourceOfFund: request.SourceOfFund,
			VaType:       vaAccountType,
		})
		if err != nil {
			request.Exception = constants.SERVICE_MERCHANT_FEE_ERROR + err.Error()
			us.Logger.With(zap.Error(err)).Error("[SERVICE_MERCHANT_FEE].error")
			if request.TransactionType != constants.TRANSTYPE_PAY_VA {
				return &service_transaction.ETransactionDTO{}, err
			}
		} else {
			request.RateFeeAmount = merchantFee.RateFeeAmount
			request.FixedFeeAmount = merchantFee.FixedFeeAmount

			if merchantFee.FeeMethod == service_merchant_fee.CheckMerchantQuotaAndFeeRes_FEE_NOW {
				request.AmountMerchantFee = merchantFee.FeeAmount
			} else {
				request.AmountMerchantFeeGpayTmp = merchantFee.FeeAmount
				request.MerchantFeeMethod = merchantFee.FeeMethod.String()
			}
		}
	}

	request.AmountFeeGpay = userFee.FeeAmount
	request.Currency = constants.VND
	request.Status = constants.TRANSACTION_STATUS_PROCESSING
	request.State = constants.TRANSACTION_STATE_INITIAL
	request.TypeWallet = constants.AmountCash

	initTran, err := us.TransactionRepository.InitTransaction(ctx, request)

	if err == nil {
		us.MqttUpdateProfile(ctx, constants.UPDATE_USER_INFO, initTran.PayerId, initTran.PayeeId)
		us.SendMqttTransactionByObject(ctx, constants.UPDATE_TRANSACTION, *initTran)
	}
	return initTran, err
}

func (us *OrderApplication) serviceTransactionCancel(ctx context.Context, dto *service_transaction.ETransactionDTO) (*service_transaction.ETransactionDTO, error) {
	cancelTran, err := us.TransactionRepository.CancelTransaction(ctx, dto)
	if err == nil && dto != nil {
		_ = us.MqttUpdateProfile(ctx, constants.UPDATE_USER_INFO, dto.PayerId, dto.PayeeId)
		_ = us.SendMqttTransactionByObject(ctx, constants.UPDATE_TRANSACTION, *cancelTran)

	}
	return cancelTran, err
}

func (us *OrderApplication) serviceTransactionPending(ctx context.Context, dto *service_transaction.ETransactionDTO) (*service_transaction.ETransactionDTO, error) {
	dto.Status = constants.TRANSACTION_STATUS_PENDING
	pendingTran, err := us.TransactionRepository.UpdateTransaction(ctx, dto)
	if err == nil {
		us.MqttUpdateProfile(ctx, constants.UPDATE_USER_INFO, pendingTran.PayerId, pendingTran.PayeeId)
		us.SendMqttTransactionByObject(ctx, constants.UPDATE_TRANSACTION, *pendingTran)
	}

	return pendingTran, err
}

func (us *OrderApplication) serviceTransactionConfirm(ctx context.Context, dto *service_transaction.ETransactionDTO) (*service_transaction.ETransactionDTO, error) {
	confirmTran, err := us.TransactionRepository.ConfirmTransaction(ctx, dto)
	if err == nil {
		us.MqttUpdateProfile(ctx, constants.UPDATE_USER_INFO, confirmTran.PayerId, confirmTran.PayeeId)
		us.SendMqttTransactionByObject(ctx, constants.UPDATE_TRANSACTION, *confirmTran)
	}
	us.notificationTransaction(ctx, *dto)

	return confirmTran, err
}

func (us *OrderApplication) notificationTransaction(ctx context.Context, dto service_transaction.ETransactionDTO) error {
	tran, _ := us.ConvertETransToDetail(ctx, dto)
	us.IPool.Submit(func() {
		switch dto.TransactionType {
		case constants.TRANSTYPE_WALLET_LIXI:
			data := make(map[string]interface{})
			data["lixi"] = dto
			data["type"] = "LIXI"

			us.CreateMessage("Gpay", dto.PayeeId, "Chúc mừng! Bạn vừa nhận được lì xì từ "+dto.Message, data)
		case constants.TRANSTYPE_WALLET_REFUND:
			ac := accounting.DefaultAccounting(" ", 0)
			ac.Thousand = "."
			_, err := us.UserRepository.CreateMessageNormal(ctx, &service_user.CreateMessageNormalRequest{
				Title:           "Hoàn tiền giao dịch thành công",
				UserId:          dto.UserReceiveRefund,
				Content:         "Bạn vừa được hoàn " + fmt.Sprint(ac.FormatMoney(tran.Amount)) + " VND từ giao dịch " + dto.TransactionSourceRefund,
				TransactionId:   tran.TransactionId,
				TransactionType: constants.TRANSTYPE_WALLET_REFUND,
			})
			if err != nil {
				us.Logger.With(zap.Error(err)).Error(constants.SERVICE_USER_ERROR)
			}
			break
		}

		switch tran.Type {
		case constants.TransactionTypeTransferGpoint:
			ac := accounting.DefaultAccounting(" ", 0)
			ac.Thousand = "."

			// notification
			us.UserRepository.CreateMessageNormal(ctx, &service_user.CreateMessageNormalRequest{
				Title:           "Bạn vừa nhận tiền thành công",
				UserId:          tran.Payee.Id,
				Content:         "Bạn vừa nhận được " + fmt.Sprint(ac.FormatMoney(tran.Amount)) + " VND từ số điện thoại ",
				TransactionId:   tran.TransactionId,
				TransactionType: tran.Type,
			})

			break
		case constants.TransactionTypeBalanceChangeDeposit:
			// notify "nap tien" success
			message := "Bạn vừa nạp thành công " + fmt.Sprint(humanize.Comma(int64(tran.Amount))) + " VNĐ"
			us.UserRepository.CreateMessageNormal(ctx, &service_user.CreateMessageNormalRequest{
				Title:           "Bạn vừa nạp thành công",
				UserId:          tran.PayeeId,
				Content:         message,
				TransactionId:   tran.TransactionId,
				TransactionType: tran.Type,
			})

			break
		case constants.TransactionTypeBalanceChangeWithdraw:
			// notify "rut tien" success
			message := "Bạn vừa rút thành công " + fmt.Sprint(humanize.Comma(int64(tran.Amount))) + " VNĐ"
			us.UserRepository.CreateMessageNormal(ctx, &service_user.CreateMessageNormalRequest{
				Title:           "Bạn vừa rút thành công",
				UserId:          tran.PayerId,
				Content:         message,
				TransactionId:   tran.TransactionId,
				TransactionType: tran.Type,
			})
			break

		}
	})
	return nil
}
