package application

import (
	"context"
	eBankGw "orders-system/domain/entities/bank_gateway"
	pb "orders-system/proto/order_system"
)

func (usecase OrderApplication) ConfirmPayByToken(ctx context.Context, req pb.ConfirmPaymentTokenRequest) (res eBankGw.VerifyOTPResponse, err error) {
	cashInVerifyOTP, err := usecase.BankServiceRepository.VerifyOTP(req.GpayBankCode, req.BankTraceId, req.OrderId, req.LinkId, req.Otp)
	if err != nil {
		return cashInVerifyOTP, err
	}

	return cashInVerifyOTP, err
}
