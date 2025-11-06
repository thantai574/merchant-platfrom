package application

import (
	"context"
	"orders-system/domain/constants"
	"orders-system/proto/service_promotion"
)

func (us *OrderApplication) servicePromotionUsed(ctx context.Context, dto *service_promotion.UseVoucherRequest) (*service_promotion.UseVoucherResponse, error) {
	if dto.UserId != "" {
		user, err := us.GetProfile(ctx, dto.UserId)
		if err == nil {
			dto.CurrentBalance = user.Balances[0].AmountAvailable - user.Balances[0].AmountFreeze
		}
	}

	res, err := us.PromotionRepository.UseVoucher(ctx, dto)

	_ = us.CreateMessageMqtt(ctx, dto.UserId, constants.MQTTEventBackground, "restart-my-voucher", "Success", false)

	return res, err
}

func (us *OrderApplication) servicePromotionCompensate(ctx context.Context, dto *service_promotion.ReverseWalletRequest) (*service_promotion.ReverseWalletRequest, error) {
	return us.PromotionRepository.ReverseWallet(ctx, dto)
}
