package test

import (
	"context"
	"github.com/stretchr/testify/mock"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	entities2 "orders-system/domain/entities/bank_gateway"
	"orders-system/proto/order_system"
	"orders-system/proto/service_merchant_fee"
	"orders-system/proto/service_transaction"
	"orders-system/proto/service_user"
	"testing"
	"time"
)

func TestOrderApplication_Lixi(t *testing.T) {
	ctx, _ := context.WithTimeout(context.TODO(), time.Second*30)
	th := NewTestOrderApplication()
	defer th.DB.Drop(ctx)
	tests := []struct {
		name      string
		id        string
		expect    string
		wantError bool
		fd        func()
	}{

		{
			name:      "test-case-2 fail KYC payee ",
			wantError: true,
			fd: func() {
				th.OrderRepository.On("CheckLuckyMoney", ctx, mock.Anything, mock.Anything).Return(nil, nil).Times(1)
				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false).Return(nil)
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id: "U",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 100000,
							AmountFreeze:    0,
						}},
					},
				}, nil).Times(2)
				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:  "user-qwe1",
							Kyc: "INACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)
				th.MerchantFeeService.On("CheckMerchantQuotaAndFee", ctx, mock.Anything).Return(&service_merchant_fee.CheckMerchantQuotaAndFeeRes{}, nil).Times(1)
				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{
					ErrorCode: "",
					Data:      []*entities2.ListLinked{},
				}, nil)
				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("UpdateTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fd()
			err := th.OrderApplication.LixiOrchestrator(ctx, &order_system.LixiRequest{
				Lixi: &order_system.Lixi{
					ID: "TEST",
				},
				OrderRequest: &order_system.OrderRequest{
					Amount:     10000,
					MerchantID: "MC",
					TransType:  constants.TRANSTYPE_WALLET_LIXI,
					ToUserID:   "U",
				},
			}, &order_system.LixiResponse{
				Lixi:  nil,
				Order: nil,
			})
			if tt.wantError == false {
				if err != nil {
					t.Errorf("want no err, but get err %v ", err)
				}
			}
			if tt.wantError == true {
				t.Log(err)
			}

		})
	}
}
