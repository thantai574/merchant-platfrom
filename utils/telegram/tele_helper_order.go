package telegram

import (
	"fmt"
	"orders-system/domain/entities"
	"orders-system/proto/service_transaction"
	"orders-system/utils/helpers"
)

func SendOrderInfo(order_dto entities.OrderEntity) string {
	orderType := order_dto.OrderType + " - " + order_dto.SubOrderType

	return fmt.Sprintf(fmt.Sprintf(`
Mã order: %v
Merchant code: %v
Loại giao dịch: %v
Trạng thái: %v 
Số tiền:  %v
Thời gian tạo:  %v
					`,
		order_dto.OrderID,
		order_dto.MerchantCode,
		orderType,
		order_dto.Status.StatusString(),
		order_dto.Amount,
		order_dto.CreatedAt.In(helpers.LocationVietNam()).Format("02-01-2006 15:04:05"),
	))
}

func SendOrderFraud(order_dto entities.OrderEntity, trans service_transaction.ETransactionDTO, fraud entities.Fraud) string {
	return fmt.Sprintf(fmt.Sprintf(`
Cảnh báo Blacklist giao dịch thẻ quốc tế
Giao dịch: %v
Số tiền :  %v
Thời gian tạo đơn hàng:  %v
Số thẻ:  %v
			%v		`,
		trans.TransactionId,
		order_dto.Amount,
		order_dto.CreatedAt.In(helpers.LocationVietNam()).Format("02-01-2006 15:04:05"),
		fraud.CardNumber,
		"",
	))
}

func SendMerchantRefund(orderDto entities.OrderEntity, eRefund service_transaction.RefundTransactionResponse, eSourceTrans service_transaction.ETransactionDTO) string {
	var typeCOnfirm = "MANUAL"
	if orderDto.Status.IsSuccess() {
		typeCOnfirm = "AUTOMATIC"
	}

	return fmt.Sprintf(fmt.Sprintf(`
REQUEST REFUND MERCHANT 
Have a request refund from  %v 
ID : %v
Amount :  %v
Service Type : %v
TransType : %v
Time :  %v
Type of confirm:  %v
			`,
		orderDto.MerchantCode,
		eRefund.RefundId,
		orderDto.Amount,
		eSourceTrans.ServiceType,
		eSourceTrans.TransactionType,
		orderDto.CreatedAt.In(helpers.LocationVietNam()).Format("02-01-2006 15:04:05"),
		typeCOnfirm,
	))
}
