package application

import (
	"context"
	"errors"
	"orders-system/proto/order_system"
)

func (us *OrderApplication) VerifyOTPLinkBank(ctx context.Context, request *order_system.VerifyOTPLinkBankRequest) (res *order_system.VerifyOTPLinkBankResponse, err error) {
	cashInVerifyOTP, err := us.BankServiceRepository.VerifyOTP(request.GpayBankCode, "", "", request.LinkId, request.Otp)
	if err != nil {
		return res, err
	}

	if !cashInVerifyOTP.ErrorCode.IsSuccess() {
		return res, errors.New(cashInVerifyOTP.Message)
	}

	return res, err
}
