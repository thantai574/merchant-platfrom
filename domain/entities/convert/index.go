package convert

import (
	"orders-system/proto/order_system"
	"orders-system/proto/service_card"
)

func FromCardInvoiceToOrderInvoiceDTO(i *service_card.InvoiceDTO) *order_system.InvoiceDTO {
	invoiceAttr := make([]*order_system.InvoiceAttribute, len(i.InvoiceAttributes))

	for index, data := range i.InvoiceAttributes {
		invoiceAttr[index] = &order_system.InvoiceAttribute{
			InvoiceId:              data.InvoiceId,
			InvoiceAttributeTypeId: data.InvoiceAttributeTypeId,
			Value:                  data.Value,
			Created:                data.Created,
		}
	}

	return &order_system.InvoiceDTO{
		TransactionId:           i.TransactionId,
		TransactionDate:         i.TransactionDate,
		Message:                 i.Message,
		IsPartialPaymentAllowed: i.IsPartialPaymentAllowed,
		Paid:                    i.Paid,
		Amount:                  i.Amount,
		InvoiceReference:        i.InvoiceReference,
		BillingCode:             i.BillingCode,
		ServiceCode:             i.ServiceCode,
		InvoiceAttributes:       invoiceAttr,
	}
}
