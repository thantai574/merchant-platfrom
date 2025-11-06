package application

import (
	"context"
	"orders-system/proto/order_system"
)

func (us *OrderApplication) IBFTInquiry(ctx context.Context, request *order_system.IBFTInquiryRequest, response *order_system.IBFTInquiryResponse) (err error) {
	getNameIBFTInfo, err := us.BankServiceRepository.IBFTInquiry(request.AccountNo, request.CardNo, request.IbftCode)
	if err != nil {
		return err
	}
	response.FullName = getNameIBFTInfo.Data.FullName
	return
}
