package application

import (
	"context"
	"orders-system/proto/order_system"
	"orders-system/utils/convert_model"
)

func (us *OrderApplication) LinkList(ctx context.Context, request *order_system.BankLinkListRequest, response *order_system.BankLinkListResponse) (err error) {
	getListLinkBank, err := us.BankServiceRepository.LinkList(request.GpayUserID)
	if err != nil {
		return err
	}

	var res []*order_system.BankLinkDetail
	for _, value := range getListLinkBank.Data {
		cv := convert_model.FromListLinkToDTO(*value)
		res = append(res, &cv)
	}
	response.LinkDetail = res
	return
}
