package application

import (
	"context"
	"encoding/json"
	"github.com/spf13/cast"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	eBank "orders-system/domain/entities/bank_gateway"
	"orders-system/proto/order_system"
	"orders-system/proto/service_transaction"
	"orders-system/utils/saga"
)

func (us *OrderApplication) NapasInitCashIn(ctx context.Context, request *order_system.NapasInitCashInRequest) (response *order_system.NapasInitCashInResponse, err error) {
	sg := saga.NewSaga("BuyCardActionAccount")

	amount := cast.ToInt64(request.Amount)
	var failReason string
	var napasCashInResp eBank.NapasCashInResponse
	var trans *service_transaction.ETransactionDTO
	orderDto := &entities.OrderEntity{}

	response = &order_system.NapasInitCashInResponse{}

	//@todo InitOrder
	err = sg.AddStep(&saga.Step{
		Name: "INIT_ORDER",
		Func: func(c context.Context) (err error) {
			serviceType := request.GetServiceType()

			if serviceType == "" {
				serviceType = constants.SERVICE_TYPE_WALLET
			}

			orderType := request.GetTransType()
			if orderType == "" {
				orderType = constants.TRANSTYPE_WALLET_CASH_IN
			}

			subOrderType := request.GetSubTransType()

			orderDto, err = us.InitOrder(ctx, &entities.OrderEntity{
				ToUserID:            request.UserId,
				OrderType:           orderType,
				Amount:              amount,
				SubOrderType:        subOrderType,
				ServiceType:         serviceType,
				SourceOfFund:        constants.SOURCE_OF_FUND_BANK_ATM,
				BankCode:            request.GpayBankCode,
				Napas:               true,
				Description:         "NAPAS CASHIN",
				ServiceCode:         constants.TRANSTYPE_WALLET_CASH_IN,
				SubscribeMerchantID: request.GetMerchantId(),
				RefID:               request.GetRefId(),
			})
			return err
		},
		CompensateFunc: func(c context.Context) (err error) {
			return err
		},
		Options: nil,
	})
	if err != nil {
		return
	}

	//todo InitTrans
	err = sg.AddStep(&saga.Step{
		Name: "INIT_TRANS",
		Func: func(c context.Context) (err error) {
			trans, err = us.serviceTransactionInit(ctx, &service_transaction.ETransactionDTO{
				OrderId:            orderDto.OrderID,
				Napas:              true,
				SourceOfFund:       orderDto.SourceOfFund,
				Message:            orderDto.Description,
				BankCode:           orderDto.BankCode,
				Amount:             orderDto.Amount,
				TransactionType:    orderDto.OrderType,
				ServiceType:        orderDto.ServiceType,
				PayeeId:            orderDto.ToUserID,
				SubTransactionType: orderDto.SubOrderType,
				BankTransactionId:  orderDto.OrderID,
				AppId:              orderDto.ServiceCode,
				RefId:              orderDto.RefID,
				MerchantID:         orderDto.SubscribeMerchantID,
			})
			if err != nil {
				return err
			}

			orderDto.TransactionID = trans.TransactionId
			return err
		},
		CompensateFunc: func(c context.Context) (err error) {
			return err
		},
		Options: nil,
	})
	if err != nil {
		return
	}

	//todo Callbank
	err = sg.AddStep(&saga.Step{
		Name: "CALL_BANK_SERVICE",
		Func: func(c context.Context) (err error) {
			dataReq := eBank.NapasCashInDataRequest{
				Amount:            amount,
				GpayTransactionID: orderDto.OrderID,
				LinkID:            request.LinkId,
				GPayUserID:        request.UserId,
				Description:       "NAPAS CASHIN",
				Channel:           "WEB",
			}

			napasCashInResp, err = us.BankServiceRepository.CashInNapas(dataReq, request.GpayBankCode, request.ClientIP)
			if err != nil {
				orderDto.TransactionID = trans.TransactionId
				failReason = err.Error()
				return err
			}

			bRes, err := json.Marshal(napasCashInResp.Data)
			if err != nil {
				return err
			}

			type DataResponse struct {
				AcsUrl string `json:"acs_url"`
			}
			var dataResponse DataResponse
			err = json.Unmarshal(bRes, &dataResponse)

			response.NapasLinkResponse = bRes
			response.Url = dataResponse.AcsUrl
			return err
		},
		CompensateFunc: func(c context.Context) (err error) {
			if trans != nil && trans.Status != constants.TRANSACTION_STATUS_FAILED && trans.Status != constants.TRANSACTION_STATUS_PENDING {
				trans.FailReason = failReason
				trans, err = us.serviceTransactionCancel(ctx, trans)
			}
			if !orderDto.Status.IsVerifying() && !orderDto.Status.IsFailed() {
				orderDto.InternalErr = failReason
				_, _ = us.FailedOrder(ctx, orderDto)
			}
			return err
		},
		Options: nil,
	})
	if err != nil {
		return
	}

	//@todo ProcessOrder
	err = sg.AddStep(&saga.Step{
		Name: "PROCESS_ORDER",
		Func: func(c context.Context) (err error) {
			orderDto.TransactionID = trans.TransactionId
			orderDto.BankTransactionId = orderDto.OrderID
			_, err = us.ProcessingOrder(ctx, orderDto)

			return err
		},
		CompensateFunc: func(c context.Context) (err error) {
			return err
		},
		Options: nil,
	})
	if err != nil {
		return
	}

	ordinator := saga.NewCoordinator(ctx, ctx, sg, us.LogSaga)
	rg := ordinator.Play()
	err = rg.ExecutionError

	return
}
