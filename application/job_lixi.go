package application

import (
	"context"
	"go.uber.org/zap/zapcore"
	"orders-system/domain/constants"
	"orders-system/proto/order_system"
	"orders-system/proto/service_user"
)

func (us *OrderApplication) JobLixi() {
	lxis, err := us.LixiRepository.FindStarted(context.Background())
	if err == nil {
		for _, lixi := range lxis {
		job_user:
			for _, phone := range lixi.UserPhones {
				user_amount := int64(0)
				if lixi.NumberScan < lixi.MaxUser {
					user_detail, err := us.UserRepository.FindUserByPhone(context.TODO(), &service_user.FindUserByPhoneRequest{
						Phone: phone,
					})
					if err == nil {
						user := user_detail.User.Id
						switch lixi.Method {
						case "RANDOM":
							rand_num, err := us.LixiRepository.FindRandomAmount(context.Background(), lixi.MinAmount, lixi.MaxAmount)
							if err == nil && rand_num != nil {
								user_amount = rand_num.Amount
								if lixi.Amount-lixi.AmountUsed-user_amount < 0 {
									break job_user
								}
								us.Logger.With(zapcore.Field{
									Key:  "lixi-data",
									Type: zapcore.ReflectType,
									Interface: order_system.OrderRequest{
										Amount:     rand_num.Amount,
										MerchantID: lixi.MerchantID,
										TransType:  constants.TRANSTYPE_WALLET_LIXI,
										ToUserID:   user,
									},
								}).Info("STARTED JobLixi ")

								err = us.LixiOrchestrator(context.Background(), &order_system.LixiRequest{
									Lixi: &order_system.Lixi{
										ID:   lixi.ID,
										Name: lixi.Name,
									},
									OrderRequest: &order_system.OrderRequest{
										Amount:     rand_num.Amount,
										MerchantID: lixi.MerchantID,
										TransType:  constants.TRANSTYPE_WALLET_LIXI,
										ToUserID:   user,
									},
								}, &order_system.LixiResponse{})
								if err == nil {
									lixi.AmountUsed = lixi.AmountUsed + user_amount
									lixi.AmountRemain = lixi.Amount - lixi.AmountUsed
									lixi.NumberScan = lixi.NumberScan + 1
								}

							}

						case "FIXED":
							user_amount = lixi.FixAmount
							if lixi.Amount-lixi.AmountUsed-user_amount < 0 {
								break job_user
							}
							us.Logger.With(zapcore.Field{
								Key:  "lixi-data",
								Type: zapcore.ReflectType,
								Interface: order_system.OrderRequest{
									Amount:     lixi.FixAmount,
									MerchantID: lixi.MerchantID,
									TransType:  constants.TRANSTYPE_WALLET_LIXI,
									ToUserID:   user,
								},
							}).Info("STARTED JobLixi ")

							err = us.LixiOrchestrator(context.Background(), &order_system.LixiRequest{
								Lixi: &order_system.Lixi{
									ID:   lixi.ID,
									Name: lixi.Name,
								},
								OrderRequest: &order_system.OrderRequest{
									Amount:     lixi.FixAmount,
									MerchantID: lixi.MerchantID,
									TransType:  constants.TRANSTYPE_WALLET_LIXI,
									ToUserID:   user,
								},
							}, &order_system.LixiResponse{})

							if err == nil {
								lixi.AmountUsed = lixi.AmountUsed + user_amount
								lixi.AmountRemain = lixi.Amount - lixi.AmountUsed
								lixi.NumberScan = lixi.NumberScan + 1
							}

						}
					}

				}

			}

			lixi.Status = "PAUSED"
			us.LixiRepository.UpdateOne(context.TODO(), lixi)

		}
	}
}
