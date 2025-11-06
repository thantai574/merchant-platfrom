package application

import (
	"context"
	"encoding/json"
	entities "orders-system/domain/entities/bank_gateway"
	pb "orders-system/proto/order_system"

	"github.com/spf13/cast"
	"go.uber.org/zap"
)

// deprecated
func (usecase OrderApplication) PGInitTrans(ctx context.Context, req *pb.PGInitTransReq, res *pb.PGInitTransRes) (err error) {
	r, err := usecase.BankServiceRepository.PGInitTrans(req.BankCode, entities.PGInitTransReqData{
		GPayTransactionID: req.OrderId,
		Description:       req.Description,
		Amount:            cast.ToString(req.Amount),
		MerchantCode:      req.MerchantCode,
		RedirectURL:       req.RedirectUrl,
	})
	if err != nil {
		usecase.Logger.Error("BankServiceRepository.PGInitTrans error", zap.Any("req", req), zap.Error(err))
		return
	}

	res.RedirectUrl = r.Data.RedirectURL
	return
}

func (usecase OrderApplication) PGInitOrder(ctx context.Context, req *pb.PGInitOrderReq, res *pb.PGInitOrderRes) (err error) {
	data := map[string]interface{}{}
	json.Unmarshal(req.Data, &data)
	r, err := usecase.BankServiceRepository.PGInitOrder(entities.PGInitTransReq{
		GPayBankCode: req.BankCode,
		IpAddress:    req.Ip,
		Data:         data,
	})
	if err != nil {
		usecase.Logger.Error("BankServiceRepository.PGInitOrder error", zap.Any("req", req), zap.Error(err))
		return
	}

	b, _ := json.Marshal(r.Data)
	res.Data = b
	return
}
