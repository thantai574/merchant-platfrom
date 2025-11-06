package application

import (
	"context"
	"orders-system/proto/order_system"
)

func (us *OrderApplication) UnLink(ctx context.Context, request *order_system.BankUnlinkRequest, response *order_system.BankUnlinkResponse) (err error) {
	linkDetail, err := us.BankServiceRepository.LinkInfo(request.LinkId)
	if err != nil {
		return err
	}

	bankCodeGpay := linkDetail.Data.GpayBankCode

	_, err = us.BankServiceRepository.UnLink(bankCodeGpay, request.LinkId)
	if err != nil {
		return err
	}

	response.Status = "Success"
	return
}
