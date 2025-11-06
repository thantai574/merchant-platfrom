package application

import (
	"context"
	"encoding/json"
	"orders-system/proto/order_system"
	pb "orders-system/proto/order_system"

	"github.com/spf13/cast"
	"go.uber.org/zap"
)

func (usecase OrderApplication) BankRetrieveOrder(ctx context.Context, req *pb.BankRetrieveOrderReq, res *pb.BankRetrieveOrderRes) (err error) {
	r, err := usecase.BankServiceRepository.RetrieveOrderStatus(req.OrderId, req.BankCode)
	if err != nil {
		usecase.Logger.Error("BankServiceRepository.RetrieveOrderStatus", zap.Any("orderID", req.OrderId), zap.Error(err))
		return
	}

	res.Success = r.ErrorCode.IsSuccess()
	res.Amount = cast.ToInt64(r.Data.Amount)
	res.BankTransactionId = r.Data.BankTransactionId
	res.ErrorCode = string(r.ErrorCode)
	res.Message = r.Message
	res.BankCode = r.Data.BankCode
	res.GpayBankCode = r.Data.GpayBankCode
	return
}

func (usecase OrderApplication) GetBanks(ctx context.Context, req *pb.GetBanksReq, res *pb.GetBanksRes) (err error) {
	r, err := usecase.BankServiceRepository.GetBanks()
	if err != nil {
		usecase.Logger.Error("BankServiceRepository.GetBanks", zap.Error(err))
		return
	}

	b, _ := json.Marshal(r.Data)
	res.Data = []*order_system.Bank{}
	json.Unmarshal(b, &res.Data)
	return
}
