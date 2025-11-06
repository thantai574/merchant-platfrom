package application

import (
	"context"
	"errors"
	"orders-system/domain/constants"
	"orders-system/domain/entities/convert"
	"orders-system/proto/order_system"
	"orders-system/proto/service_card"
)

func (us *OrderApplication) CheckBill(ctx context.Context, request *order_system.CheckBillRequest, response *order_system.CheckBillResponse) (err error) {
	subTransactionType := ""
	findVendorByServiceCode, err := us.ServiceCardRepository.FindVendorByCode(ctx, &service_card.FindVendorByCodeReq{
		ServiceCode: request.ServiceCode,
	})

	if err != nil {
		return
	}

	switch findVendorByServiceCode.Vendor.Type {
	case "electric":
		subTransactionType = constants.SUB_TRANSTYPE_WALLET_PAY_BILL_ELECTRIC
	case "water":
		subTransactionType = constants.SUB_TRANSTYPE_WALLET_PAY_BILL_WATTER
	case "home_credit":
		subTransactionType = constants.SUB_TRANSTYPE_WALLET_PAY_BILL_LOAN
	case "tv":
		subTransactionType = constants.SUB_TRANSTYPE_WALLET_PAY_BILL_TV
	case "internet":
		subTransactionType = constants.SUB_TRANSTYPE_WALLET_PAY_BILL_INTERNET
	case "telephone":
		subTransactionType = constants.SUB_TRANSTYPE_WALLET_PAY_BILL_TELEPHONE
	}

	r, err := us.ServiceCardRepository.InvoiceCheck(ctx, &service_card.InvoiceCheckReq{
		ServiceCode:       request.ServiceCode,
		CustomerReference: request.BillingCode,
		SubTransType:      subTransactionType,
		AreaCode:          request.AreaCode,
	})

	if err != nil {
		return errors.New("Mã khách hàng không tồn tại hoặc chưa phát sinh hóa đơn thanh toán. Vui lòng kiểm tra lại hoặc thanh toán vào lúc khác")
	}

	response.Invoice = convert.FromCardInvoiceToOrderInvoiceDTO(r.Invoice)
	response.Invoice.ProviderName = findVendorByServiceCode.Vendor.Name
	response.Status = r.Status
	return
}
