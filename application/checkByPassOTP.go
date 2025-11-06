package application

import (
	"context"
	"github.com/spf13/cast"
	entities "orders-system/domain/entities/bank_gateway"
	"orders-system/proto/order_system"
)

func (us *OrderApplication) CheckByPassOTP(ctx context.Context, req *order_system.CheckByPassOTPRequest) (response *order_system.CheckByPassOTPResponse, err error) {
	response = new(order_system.CheckByPassOTPResponse)

	byPassOTPResp, err := us.BankServiceRepository.CheckByPassOTP(entities.CheckByPassOTPDataReq{
		Amount:       cast.ToString(req.Amount),
		ApiName:      req.ApiName,
		GpayBankCode: req.GpayBankCode,
		GpayUserId:   req.UserId,
	})

	if err != nil {
		return response, err
	}

	if byPassOTPResp.ErrorCode.IsSuccess() {
		response.IsByPassOTP = true
		response.Description = "ByPassOTP"
	}
	if byPassOTPResp.ErrorCode.IsNeedToEnterOTP() {
		response.IsByPassOTP = false
		response.Description = "Not ByPassOTP"
	}

	return response, err
}
