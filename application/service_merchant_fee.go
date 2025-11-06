package application

import (
	"context"
	pb "orders-system/proto/order_system"
	"orders-system/proto/service_merchant_fee"

	"go.uber.org/zap"
)

// for other services call
func (application *OrderApplication) CheckMerchantQuotaAndFee(ctx context.Context,
	req *pb.CheckMerchantQuotaAndFeeReq, res *pb.CheckMerchantQuotaAndFeeRes) (err error) {
	merchantFee, err := application.MerchantFeeRepository.CheckMerchantQuotaAndFee(ctx, &service_merchant_fee.CheckMerchantQuotaAndFeeReq{
		Amount:       req.Amount,
		ServiceType:  req.ServiceType,
		TransType:    req.TransType,
		SubTransType: req.SubTransType,
		MerchantId:   req.MerchantId,
		SourceOfFund: req.SourceOfFund,
		VaType:       service_merchant_fee.CheckMerchantQuotaAndFeeReq_VAType(req.VaType),
	})
	if err != nil {
		application.Logger.Error("MerchantFeeRepository.CheckMerchantQuotaAndFee error", zap.Any("req", req), zap.Error(err))
		return
	}

	res.FeeAmount = merchantFee.FeeAmount
	res.FeeMethod = pb.CheckMerchantQuotaAndFeeRes_FeeMethod(merchantFee.FeeMethod)
	res.FixedFeeAmount = merchantFee.FixedFeeAmount
	res.RateFeeAmount = merchantFee.RateFeeAmount
	return
}
