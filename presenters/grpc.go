package presenters

import (
	"context"
	"errors"
	"orders-system/application"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	"orders-system/proto/order_system"
	"orders-system/utils/helpers"

	context2 "golang.org/x/net/context"
)

type OrderSystemGRPC struct {
	OrderApplication *application.OrderApplication
}

func (o *OrderSystemGRPC) TransferWallet(c context2.Context, request *order_system.TransferWalletRequest) (response *order_system.TransferWalletResponse, err error) {
	response, err = o.OrderApplication.TransferGpayWallet(c, request)
	if err != nil {
		return response, err
	}

	return
}

func (o *OrderSystemGRPC) GetOrderByMerchant(c context2.Context, req *order_system.GetOrderByMerchantReq) (response *order_system.GetOrderByMerchantRes, err error) {
	response, err = o.OrderApplication.GetOrderByMerchant(c, req)
	if err != nil {
		return response, err
	}

	return
}

func (o *OrderSystemGRPC) InitInternationalLinkBank(c context2.Context, request *order_system.LinkRequest) (response *order_system.LinkResponse, err error) {
	response, err = o.OrderApplication.LinkBank(c, request)
	if err != nil {
		return response, err
	}

	return
}

func (o *OrderSystemGRPC) PayInternationalCard(c context2.Context, request *order_system.PayInternationalCardRequest) (response *order_system.PayInternationalCardResponse, err error) {
	response, err = o.OrderApplication.PayInternationalCard(c, request)
	if err != nil {
		return response, err
	}

	return
}

func (o *OrderSystemGRPC) VerifyOTPLinkBank(c context2.Context, request *order_system.VerifyOTPLinkBankRequest) (response *order_system.VerifyOTPLinkBankResponse, err error) {
	response, err = o.OrderApplication.VerifyOTPLinkBank(c, request)
	if err != nil {
		return response, err
	}

	return new(order_system.VerifyOTPLinkBankResponse), err
}

func (o *OrderSystemGRPC) Link(c context2.Context, request *order_system.LinkRequest) (response *order_system.LinkResponse, err error) {
	response, err = o.OrderApplication.LinkBank(c, request)
	if err != nil {
		return response, err
	}

	return
}

func (o *OrderSystemGRPC) CheckByPassOTP(c context2.Context, request *order_system.CheckByPassOTPRequest) (response *order_system.CheckByPassOTPResponse, err error) {
	response, err = o.OrderApplication.CheckByPassOTP(c, request)
	if err != nil {
		return response, err
	}

	return
}

func (o *OrderSystemGRPC) CancelRefundTransaction(c context2.Context, req *order_system.CancelRefundTransactionReq) (response *order_system.CancelRefundTransactionRes, err error) {
	response = new(order_system.CancelRefundTransactionRes)

	_, err = o.OrderApplication.CancelRefundTrans(c, req)
	if err != nil {
		return response, err
	}

	return
}

func (o *OrderSystemGRPC) InitRefundTransaction(c context2.Context, req *order_system.RefundTransactionReq) (response *order_system.RefundTransactionRes, err error) {
	response = new(order_system.RefundTransactionRes)

	res, err := o.OrderApplication.InitRefundTrans(c, req)
	if err != nil {
		return response, err
	}

	response.OrderEntity = res.ConvertToProto()
	return
}

func (o *OrderSystemGRPC) ConfirmRefundTransaction(c context2.Context, req *order_system.ConfirmRefundTransactionReq) (response *order_system.ConfirmRefundTransactionRes, err error) {
	response = new(order_system.ConfirmRefundTransactionRes)

	res, err := o.OrderApplication.ConfirmRefundTrans(c, req)
	if err != nil {
		return response, err
	}

	response.OrderEntity = res.ConvertToProto()
	return
}

func (o *OrderSystemGRPC) NapasLink(c context2.Context, request *order_system.NapasInitLinkRequest) (response *order_system.NapasInitLinkResponse, err error) {
	response, err = o.OrderApplication.NapasInitLink(c, request)
	if err != nil {
		return response, err
	}
	return
}

func (o *OrderSystemGRPC) NapasInitCashIn(c context2.Context, request *order_system.NapasInitCashInRequest) (response *order_system.NapasInitCashInResponse, err error) {
	response, err = o.OrderApplication.NapasInitCashIn(c, request)
	if err != nil {
		return response, err
	}
	return
}

func (o *OrderSystemGRPC) ReOpenVA(c context2.Context, request *order_system.ReOpenVARequest) (response *order_system.ReOpenVAResponse, err error) {
	response = new(order_system.ReOpenVAResponse)
	err = o.OrderApplication.ActionReOpenAccountVA(c, request, response)
	return
}

func (o *OrderSystemGRPC) CreateVA(c context2.Context, request *order_system.CreateVARequest) (response *order_system.CreateVAResponse, err error) {
	if request.MerchantId == "" {
		return response, errors.New("MerchantId is required")
	}

	response = new(order_system.CreateVAResponse)
	err = o.OrderApplication.ActionCreateVA(c, request, response)
	return
}

func (o *OrderSystemGRPC) UpdateVA(c context2.Context, request *order_system.UpdateVARequest) (response *order_system.UpdateVAResponse, err error) {
	response = new(order_system.UpdateVAResponse)
	err = o.OrderApplication.ActionUpdateVA(c, request, response)
	return
}

func (o *OrderSystemGRPC) CloseVA(c context2.Context, request *order_system.CloseVARequest) (response *order_system.CloseVAResponse, err error) {
	response = new(order_system.CloseVAResponse)
	err = o.OrderApplication.ActionCloseVA(c, request, response)
	return
}

func (o *OrderSystemGRPC) DetailVA(c context2.Context, request *order_system.DetailVARequest) (response *order_system.DetailVAResponse, err error) {
	response = new(order_system.DetailVAResponse)
	err = o.OrderApplication.ActionGetDetailVA(c, request, response)
	return

}

func (o *OrderSystemGRPC) InitPaymentToken(c context2.Context, request *order_system.InitPaymentTokenRequest) (response *order_system.InitPaymentTokenResponse, err error) {
	response = new(order_system.InitPaymentTokenResponse)
	_, _, err = o.OrderApplication.InitPayByToken(c, request, response)
	return
}

func (o *OrderSystemGRPC) CancelOrder(ctx context2.Context, request *order_system.CancelOrderRequest) (response *order_system.CancelOrderResponse, err error) {
	if request.OrderId == "" {
		return response, errors.New("OrderId is required")
	}

	response = new(order_system.CancelOrderResponse)
	_, err = o.OrderApplication.OrderCancel(ctx, request, response)
	return
}

func (o *OrderSystemGRPC) UpdateCreditPayment(c context2.Context, request *order_system.UpdateCreditPaymentRequest) (response *order_system.UpdateCreditPaymentResponse, err error) {
	if request.OrderId == "" {
		return response, errors.New("OrderId is required")
	}

	response = new(order_system.UpdateCreditPaymentResponse)
	_, err = o.OrderApplication.UpdateCreditPaymentOrder(c, request, response)
	return

}

func (o *OrderSystemGRPC) ProcessCreditPayment(c context2.Context, request *order_system.ProcessCreditPaymentRequest) (response *order_system.ProcessCreditPaymentResponse, err error) {
	if request.OrderId == "" {
		return response, errors.New("OrderId is required")
	}
	if len(request.CardNumber) < 6 {
		return response, errors.New("Thẻ không hợp lệ")
	}

	response = new(order_system.ProcessCreditPaymentResponse)
	_, err = o.OrderApplication.ProcessCreditPayment(c, request, response)

	return
}

func (o *OrderSystemGRPC) InitOrder(c context2.Context, request *order_system.InitOrderRequest) (response *order_system.InitOrderResponse, err error) {
	response = new(order_system.InitOrderResponse)

	if request.OrderRequest.ExpireTime < 0 {
		return response, errors.New("Invalid expire time")
	}

	orderDto := &entities.OrderEntity{
		ServiceID:                request.OrderRequest.ServiceID,
		RefID:                    request.OrderRequest.RefID,
		UserID:                   request.OrderRequest.UserID,
		SubscribeMerchantID:      request.OrderRequest.MerchantID,
		ServiceType:              request.OrderRequest.ServiceType,
		OrderType:                request.OrderRequest.TransType,
		SubOrderType:             request.OrderRequest.SubTransType,
		Amount:                   request.OrderRequest.Amount,
		Quantity:                 request.OrderRequest.Quantity,
		SourceOfFund:             request.OrderRequest.SourceOfFund,
		VoucherCode:              request.OrderRequest.VoucherCode,
		DeviceID:                 request.OrderRequest.DeviceID,
		BankCode:                 request.OrderRequest.BankCode,
		GPayBankCode:             request.OrderRequest.GPayBankCode,
		PhoneTopUp:               request.OrderRequest.PhoneTopup,
		ToUserID:                 request.OrderRequest.ToUserID,
		MerchantCode:             request.OrderRequest.MerchantCode,
		CardNumber:               request.OrderRequest.CardNumber,
		AccountNo:                request.OrderRequest.AccountNo,
		Napas:                    request.OrderRequest.Napas,
		BankTransactionId:        request.OrderRequest.BankTransactionId,
		AmountMerchantFee:        request.OrderRequest.AmountMerchantFee,
		AmountMerchantFeeGpayTmp: request.OrderRequest.AmountMerchantFeeGpayTmp,
		MerchantCategoryCode:     request.OrderRequest.MerchantCategoryCode,
		MerchantCategoryType:     request.OrderRequest.MerchantCategoryType,
		ServiceCode:              request.OrderRequest.ServiceCode,
		MerchantTypeWallet:       request.OrderRequest.MerchantTypeWallet,
		FixedFeeAmount:           request.OrderRequest.FixedFeeAmount,
		RateFeeAmount:            request.OrderRequest.RateFeeAmount,
		ExpireTime:               request.OrderRequest.ExpireTime,
		MerchantFeeMethod:        request.OrderRequest.MerchantFeeMethod,
		Description:              request.OrderRequest.Description,
		CustomerId:               request.OrderRequest.CustomerId,
		Metadata:                 request.OrderRequest.Metadata,
	}

	res, err := o.OrderApplication.InitOrderAndTrans(c, orderDto)
	if err != nil {
		return response, err
	}

	response.OrderEntity = res.ConvertToProto()
	return response, err
}

func (o *OrderSystemGRPC) ConfirmOrder(c context2.Context, request *order_system.ConfirmOrderRequest) (response *order_system.ConfirmOrderResponse, err error) {
	if request.OrderId == "" {
		return response, errors.New("OrderId is required")
	}

	response = new(order_system.ConfirmOrderResponse)
	if request.OrderRequest != nil {
		if request.OrderRequest.SourceOfFund == constants.SOURCE_OF_FUND_BANK_ATM {
			_, err = o.OrderApplication.ConfirmOrderWithToken(c, request, response)
		} else {
			_, err = o.OrderApplication.ConfirmOrder(c, request, response)
		}
	} else {
		_, err = o.OrderApplication.ConfirmOrder(c, request, response)
	}

	return
}

func (o *OrderSystemGRPC) GetDetailOrder(c context2.Context, request *order_system.GetDetailOrderRequest) (response *order_system.GetDetailOrderResponse, err error) {
	if request.OrderId == "" {
		return response, errors.New("OrderId is required")
	}

	response = new(order_system.GetDetailOrderResponse)
	getOrderByIdRes, err := o.OrderApplication.GetDetailOrder(c, request.OrderId)

	if err != nil {
		return response, err
	}
	response.OrderEntity = getOrderByIdRes.ConvertToProto()
	return response, err
}

func (o *OrderSystemGRPC) GetDetailOrderByMerchantOrderId(c context2.Context,
	request *order_system.GetDetailOrderRequest) (response *order_system.GetDetailOrderResponse, err error) {
	if request.RefId == "" || request.MerchantId == "" {
		return response, errors.New("Invalid request")
	}

	response = new(order_system.GetDetailOrderResponse)
	getOrderByIdRes, err := o.OrderApplication.GetDetailOrderByMerchantOrderId(c, request)

	if err != nil {
		return response, err
	}
	response.OrderEntity = getOrderByIdRes.ConvertToProto()
	return response, err
}

// nap tien merchant
func (o *OrderSystemGRPC) DepositMerchant(c context2.Context, request *order_system.DepositMerchantRequest) (response *order_system.DepositMerchantResponse, err error) {
	if !helpers.IsStringSliceContains([]string{constants.AmountWallet, constants.AmountCollectionPay}, request.BankAccountType) { // must be amount_wallet, amount_collection_pay
		return response, errors.New("Bank account type is invalid")
	}

	response = new(order_system.DepositMerchantResponse)
	_, err = o.OrderApplication.DepositMerchant(c, request)
	return
}

// rut tien merchant
func (o *OrderSystemGRPC) WithrawMerchant(c context2.Context, request *order_system.WithrawMerchantRequest) (response *order_system.WithrawMerchantResponse, err error) {
	if !helpers.IsStringSliceContains([]string{constants.AmountWallet, constants.AmountPaymentGateway, constants.AmountCollectionPay, constants.AmountFee}, request.BankAccountType) {
		return response, errors.New("Bank account type is invalid")
	}

	response = new(order_system.WithrawMerchantResponse)
	_, err = o.OrderApplication.WithdrawMerchant(c, request)
	return
}

func (o *OrderSystemGRPC) PaymentMerchant(c context2.Context, request *order_system.PaymentMerchantRequest) (response *order_system.PaymentMerchantResponse, err error) {
	response = new(order_system.PaymentMerchantResponse)
	if request.OrderRequest.SourceOfFund == constants.SOURCE_OF_FUND_BANK_ATM {
		_, err = o.OrderApplication.StaticQRWithToken(c, request, response)
	} else {
		_, err = o.OrderApplication.StaticQR(c, request, response)
	}
	return
}

func (o *OrderSystemGRPC) CashInNapas(c context2.Context, request *order_system.CashInNapasRequest) (response *order_system.CashInNapasResponse, err error) {
	response = new(order_system.CashInNapasResponse)
	_, err = o.OrderApplication.CashInNAPAS(c, request, response)
	return
}

func (o *OrderSystemGRPC) CheckBill(c context2.Context, request *order_system.CheckBillRequest) (response *order_system.CheckBillResponse, err error) {
	response = new(order_system.CheckBillResponse)
	err = o.OrderApplication.CheckBill(c, request, response)
	return
}

func (o *OrderSystemGRPC) PayBill(c context2.Context, request *order_system.OrderPayBillRequest) (response *order_system.OrderPayBillResponse, err error) {
	response = new(order_system.OrderPayBillResponse)
	if request.OrderRequest.SourceOfFund == constants.SOURCE_OF_FUND_BANK_ATM {
		_, err = o.OrderApplication.PaidBillWithToken(c, request, response)
	} else {
		_, err = o.OrderApplication.PayBill(c, request, response)
	}
	return
}

func (o *OrderSystemGRPC) IBFTInquiry(c context2.Context, request *order_system.IBFTInquiryRequest) (response *order_system.IBFTInquiryResponse, err error) {
	response = new(order_system.IBFTInquiryResponse)
	err = o.OrderApplication.IBFTInquiry(c, request, response)
	return
}

func (o *OrderSystemGRPC) IBFTTransfer(c context2.Context, request *order_system.IBFTTransferRequest) (response *order_system.IBFTTransferResponse, err error) {
	response = new(order_system.IBFTTransferResponse)
	_, err = o.OrderApplication.IBFTransfer(c, request, response)
	return
}

func (o *OrderSystemGRPC) CashOut(c context.Context, request *order_system.CashOutRequest) (response *order_system.CashOutResponse, err error) {
	response = new(order_system.CashOutResponse)
	_, err = o.OrderApplication.CashOut(c, request, response)
	return
}

func (o *OrderSystemGRPC) BankUnlink(c context2.Context, request *order_system.BankUnlinkRequest) (response *order_system.BankUnlinkResponse, err error) {
	response = new(order_system.BankUnlinkResponse)
	err = o.OrderApplication.UnLink(c, request, response)
	return
}

func (o *OrderSystemGRPC) BankLinkList(c context2.Context, request *order_system.BankLinkListRequest) (response *order_system.BankLinkListResponse, err error) {
	response = new(order_system.BankLinkListResponse)
	err = o.OrderApplication.LinkList(c, request, response)
	return
}

func (o *OrderSystemGRPC) BankLinkInfo(c context2.Context, request *order_system.BankLinkInfoRequest) (*order_system.BankLinkInfoResponse, error) {
	panic("implement me")
}

func (o *OrderSystemGRPC) Lixi(c context2.Context, request *order_system.LixiRequest) (*order_system.LixiResponse, error) {
	panic("implement me")
}

func (o *OrderSystemGRPC) CashIn(c context2.Context, request *order_system.OrderCashInRequest) (response *order_system.OrderCashInResponse, err error) {
	response = new(order_system.OrderCashInResponse)
	_, err = o.OrderApplication.CashIn(c, request, response)
	return
}

func (o *OrderSystemGRPC) CashOTP(c context2.Context, request *order_system.OrderCashOTPRequest) (response *order_system.OrderCashOTPResponse, err error) {
	response = new(order_system.OrderCashOTPResponse)
	_, err = o.OrderApplication.VerifyOTP(c, request, response)
	return
}

func (o *OrderSystemGRPC) FundWallet2Bank(c context.Context, request *order_system.FundWallet2BankRequest) (response *order_system.FundWallet2BankResponse, err error) {
	response = new(order_system.FundWallet2BankResponse)
	_, err = o.OrderApplication.FundWal2Bank(c, request, response)
	return
}

func (o *OrderSystemGRPC) FundWallet2Wallet(c context.Context, request *order_system.FundWallet2WalletRequest) (response *order_system.FundWallet2WalletResponse, err error) {
	response = new(order_system.FundWallet2WalletResponse)
	_, err = o.OrderApplication.FundWal2Wal(c, request, response)
	return
}

func (o *OrderSystemGRPC) TopUp(c context.Context, request *order_system.OrderTopUpRequest) (response *order_system.OrderTopUpResponse, err error) {
	response = new(order_system.OrderTopUpResponse)
	if request.OrderRequest.SourceOfFund == constants.SOURCE_OF_FUND_BANK_ATM {
		_, err = o.OrderApplication.TopUpWithToken(c, request, response)
	} else {
		_, err = o.OrderApplication.TopUp(c, request, response)
	}
	return
}

func (o *OrderSystemGRPC) PayBillZota(c context.Context, request *order_system.OrderPayBillRequest) (response *order_system.OrderPayBillResponse, err error) {
	response = new(order_system.OrderPayBillResponse)
	_, err = o.OrderApplication.PayBill(c, request, response)
	return
}

func NewOrderSystemGRPC(orderApplication *application.OrderApplication) *OrderSystemGRPC {
	return &OrderSystemGRPC{OrderApplication: orderApplication}
}

func (o *OrderSystemGRPC) Call(context.Context, *order_system.Request) (rs *order_system.Response, e error) {
	return
}

func (o *OrderSystemGRPC) BuyCard(c context.Context, i *order_system.OrderBuyCardRequest) (r *order_system.OrderBuyCardResponse, err error) {
	r = new(order_system.OrderBuyCardResponse)
	if i.OrderRequest.SourceOfFund == constants.SOURCE_OF_FUND_BANK_ATM {
		_, err = o.OrderApplication.BuyCardWithToken(c, i, r)
	} else {
		_, err = o.OrderApplication.BuyCard(c, i, r)
	}
	return
}

// deprecated
func (o *OrderSystemGRPC) PGInitTrans(c context.Context, i *order_system.PGInitTransReq) (r *order_system.PGInitTransRes, err error) {
	r = &order_system.PGInitTransRes{}
	err = o.OrderApplication.PGInitTrans(c, i, r)
	return
}

func (o *OrderSystemGRPC) PGInitOrder(c context.Context, i *order_system.PGInitOrderReq) (r *order_system.PGInitOrderRes, err error) {
	r = &order_system.PGInitOrderRes{}
	err = o.OrderApplication.PGInitOrder(c, i, r)
	return
}

func (o *OrderSystemGRPC) BankRetrieveOrder(c context.Context, i *order_system.BankRetrieveOrderReq) (r *order_system.BankRetrieveOrderRes, err error) {
	r = &order_system.BankRetrieveOrderRes{}
	err = o.OrderApplication.BankRetrieveOrder(c, i, r)
	return
}

func (o *OrderSystemGRPC) RefundTransaction(c context2.Context, req *order_system.RefundTransactionReq) (response *order_system.RefundTransactionRes, err error) {
	if req.Amount <= 0 {
		return response, errors.New("Số tiền không hợp lệ")
	}

	refundResponse, err := o.OrderApplication.InitRefundTrans(c, req)
	if err != nil {
		return response, err
	}

	response = new(order_system.RefundTransactionRes)
	response.OrderEntity = refundResponse.ConvertToProto()
	return
}

func (o *OrderSystemGRPC) GetBanks(c context.Context, i *order_system.GetBanksReq) (r *order_system.GetBanksRes, err error) {
	r = &order_system.GetBanksRes{}
	err = o.OrderApplication.GetBanks(c, i, r)
	return
}

func (o *OrderSystemGRPC) CheckMerchantQuotaAndFee(c context.Context,
	i *order_system.CheckMerchantQuotaAndFeeReq) (r *order_system.CheckMerchantQuotaAndFeeRes, err error) {
	r = &order_system.CheckMerchantQuotaAndFeeRes{}
	err = o.OrderApplication.CheckMerchantQuotaAndFee(c, i, r)
	return
}

func (o *OrderSystemGRPC) UpdateOrder(ctx context2.Context, request *order_system.UpdateOrderRequest) (response *order_system.UpdateOrderResponse, err error) {
	if request.OrderRequest.OrderId == "" {
		return response, errors.New("OrderId is required")
	}

	response = new(order_system.UpdateOrderResponse)
	err = o.OrderApplication.OrderUpdate(ctx, request, response)
	return
}
