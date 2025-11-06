package convert_model

import (
	entities "orders-system/domain/entities/bank_gateway"
	"orders-system/proto/order_system"
)

func FromListLinkToDTO(i entities.ListLinked) order_system.BankLinkDetail {
	return order_system.BankLinkDetail{
		LinkId:            i.LinkID,
		CardNumber:        i.CardNumber,
		AccountNumber:     i.AccountNumber,
		FullName:          i.FullName,
		GpayUserID:        i.GpayUserID,
		PhoneNumber:       i.PhoneNumber,
		Provider:          i.Provider,
		BankCode:          i.BankCode,
		GPayBankCode:      i.GpayBankCode,
		Status:            i.Status,
		BankLogo:          i.Logo,
		BankPublishName:   i.BankPublishName,
		UrlDirectLinkBank: i.UrlDirectLinkBank,
		UrlDeposit:        i.UrlDeposit,
		BackgroundSmall:   i.BackgroundSmall,
		BackgroundLarge:   i.BackgroundLarge,
	}
}
