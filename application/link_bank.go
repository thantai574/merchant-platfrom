package application

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	eBankGw "orders-system/domain/entities/bank_gateway"
	"orders-system/proto/order_system"
	"orders-system/proto/service_transaction"
	"orders-system/utils/saga"
)

func (us *OrderApplication) LinkBank(ctx context.Context, req *order_system.LinkRequest) (*order_system.LinkResponse, error) {

	// check black list thẻ quốc tế
	if req.GetCvv() != "" {
		//todo check fraud
		checkFraudRes, err := us.IFraud.GetFraud(req.GetCardNumber())
		if err != nil {
			us.Logger.With(zap.Error(err)).Error("err_get_fraud")
			return nil, errors.New("Hệ thống chưa hỗ trợ liên kết thẻ trên")
		}
		if checkFraudRes.Status.IsFail() {
			return nil, errors.New("Hệ thống chưa hỗ trợ liên kết thẻ trên")
		}
	}

	linkBankResponse, err := us.BankServiceRepository.Link(eBankGw.LinkRequestData{
		GPayUserInfo: struct {
			CustomerId  string `json:"customer_id"`
			Gender      string `json:"gender,omitempty"`
			PhoneNumber string `json:"phone_number"`
			FullName    string `json:"full_name,omitempty" `

			District  string `json:"district"`
			City      string `json:"city"`
			Email     string `json:"email"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Address   string `json:"address"`
		}{
			CustomerId:  req.IdentifyNumber,
			Gender:      req.Gender,
			PhoneNumber: req.PhoneNumber,
			FullName:    req.CardHolderName,
			District:    req.Distric,
			City:        req.City,
			Email:       req.Email,
			FirstName:   req.FirstName,
			LastName:    req.LastName,
			Address:     req.Address,
		},
		GpayUserId:  req.UserId,
		ReturnUrl:   req.ReturnUrl,
		CancelUrl:   req.CancelUrl,
		Description: "Link bank " + req.GpayBankCode,
		CustomerCardInfo: struct {
			CardNumber     string `json:"card_number"`
			CardHolderName string `json:"card_holder_name"`
			Cvv            string `json:"cvv,omitempty"`
			IssueDate      string `json:"issue_date"`
			ExpireDate     string `json:"expire_date"`
		}{
			CardNumber:     req.CardNumber,
			CardHolderName: req.CardHolderName,
			Cvv:            req.Cvv,
			IssueDate:      req.IssueDate,
			ExpireDate:     req.ExpireDate,
		},
	}, req.GpayBankCode, "core.ip")
	if err != nil {
		return nil, err
	}

	bRes, err := json.Marshal(linkBankResponse.DataLinkResp)
	type DataResponse struct {
		AcsUrl string `json:"acs_url"`
	}
	var dataResponse DataResponse
	err = json.Unmarshal(bRes, &dataResponse)

	return &order_system.LinkResponse{
		LinkResponse: bRes,
		Url:          dataResponse.AcsUrl,
	}, err
}

func (us *OrderApplication) NapasInitLink(ctx context.Context, request *order_system.NapasInitLinkRequest) (response *order_system.NapasInitLinkResponse, err error) {
	sg := saga.NewSaga("NAPAS LINK")

	amount := cast.ToInt64(request.Amount)
	orderDto := &entities.OrderEntity{}
	var trans *service_transaction.ETransactionDTO
	var linkResp eBankGw.LinkResponse
	var failReason string
	response = &order_system.NapasInitLinkResponse{}

	//@todo InitOrder
	err = sg.AddStep(&saga.Step{
		Name: "INIT_ORDER",
		Func: func(c context.Context) (err error) {
			orderType := request.GetTransType()
			if orderType == "" {
				orderType = constants.TRANSTYPE_WALLET_CASH_IN
			}

			orderDto, err = us.InitOrder(ctx, &entities.OrderEntity{
				ToUserID:     request.UserId,
				OrderType:    orderType,
				Amount:       amount,
				SubOrderType: request.GetSubTransType(),
				SourceOfFund: constants.SOURCE_OF_FUND_BANK_ATM,
				BankCode:     request.GpayBankCode,
				Napas:        true,
				Description:  "NAPAS LINK",
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
				BankCode:           request.GpayBankCode,
				Amount:             orderDto.Amount,
				ServiceType:        constants.SERVICE_TYPE_WALLET,
				TransactionType:    orderDto.OrderType,
				PayeeId:            orderDto.ToUserID,
				SubTransactionType: orderDto.SubOrderType,
				BankTransactionId:  orderDto.OrderID,
				AppId:              constants.TRANSTYPE_WALLET_CASH_IN,
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

	//todo Callbank
	err = sg.AddStep(&saga.Step{
		Name: "CALL_BANK_SERVICE",
		Func: func(c context.Context) (err error) {
			dataReq := eBankGw.LinkRequestData{
				GpayTransactionId: orderDto.OrderID,
				GpayUserId:        request.UserId,
				Amount:            amount,
				Description:       "NAPAS link",
			}
			linkResp, err = us.BankServiceRepository.Link(dataReq, request.GpayBankCode, request.ClientIP)
			if err != nil {
				failReason = err.Error()
				orderDto.TransactionID = trans.TransactionId
				return err
			}

			bRes, err := json.Marshal(linkResp.DataLinkResp)
			if err != nil {
				return err
			}
			response.NapasInitLinkResponse = bRes
			return err
		},
		CompensateFunc: func(c context.Context) (err error) {
			if trans != nil && trans.Status != constants.TRANSACTION_STATUS_FAILED && trans.Status != constants.TRANSACTION_STATUS_PENDING {
				trans.FailReason = failReason
				trans, err = us.serviceTransactionCancel(ctx, trans)
			}
			if !orderDto.Status.IsVerifying() && !orderDto.Status.IsFailed() {
				orderDto.InternalErr = failReason
				us.FailedOrder(ctx, orderDto)
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
