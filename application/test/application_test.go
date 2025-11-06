package test

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	entities2 "orders-system/domain/entities/bank_gateway"
	"orders-system/domain/request_params"
	"orders-system/domain/value_objects"
	"orders-system/proto/order_system"
	"orders-system/proto/service_card"
	"orders-system/proto/service_merchant_fee"
	"orders-system/proto/service_promotion"
	"orders-system/proto/service_transaction"
	"orders-system/proto/service_user"
	"orders-system/utils/context_grpc"
	"orders-system/utils/helpers"
	"testing"
	"time"
)

func TestOrderApplication_CreateOrder(t *testing.T) {
	ctx := context.TODO()
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
			name: "test-case-1",
			fd: func() {
				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything, mock.Anything).Return(nil)

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fd()
			order, err := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{})

			if tt.wantError == false {
				if err != nil {
					assert.Error(t, err)
				}

				if !assert.Equal(t, order.Status.IsPending(), true) {
					assert.Error(t, fmt.Errorf("IsSuccess error "))
				}

				if order.OrderID == "" {
					assert.Error(t, fmt.Errorf("cannot find order "))
				}

			} else {
				assert.NotEqual(t, err, nil)
			}

		})
	}
}

func TestOrderApplication_UpdateStatusOrder(t *testing.T) {
	th := NewTestOrderApplication()
	ctx := context.TODO()
	defer th.DB.Drop(ctx)
	tests := []struct {
		name         string
		id           string
		expect       order_system.OrderStatus
		expectString string
		wantError    bool
		status       order_system.OrderStatus
		fd           func()
	}{
		{
			name:         "test-case-change-processing",
			expect:       order_system.OrderStatus_ORDER_PROCESSING,
			expectString: "ORDER_PROCESSING",
			status:       order_system.OrderStatus_ORDER_PROCESSING,
			fd: func() {
				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)
			},
		},

		{
			name:         "test-case-change-success",
			expect:       order_system.OrderStatus_ORDER_SUCCESS,
			expectString: "ORDER_SUCCESS",
			status:       order_system.OrderStatus_ORDER_SUCCESS,
			fd: func() {
				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)
			},
		},

		{
			name:         "test-case-change-cancel",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)
			},
		},
	}
	for _, tt := range tests {
		tt.fd()
		t.Run(tt.name, func(t *testing.T) {
			order, err := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{})

			switch tt.status {
			case order_system.OrderStatus_ORDER_PROCESSING:
				th.OrderApplication.ProcessingOrder(ctx, order)
			case order_system.OrderStatus_ORDER_CANCEL:
				th.OrderApplication.CancelOrder(ctx, order)
			case order_system.OrderStatus_ORDER_SUCCESS:
				th.OrderApplication.SuccessOrder(ctx, order)
			case order_system.OrderStatus_ORDER_FAILED:
				th.OrderApplication.FailedOrder(ctx, order)
			}

			if tt.wantError == false {
				if err != nil {
					assert.Error(t, err)
				}

				if !assert.Equal(t, order.Status.StatusOrderProto(), tt.expect) || !assert.Equal(t, order.Status.StatusString(), tt.expectString) {
					assert.Error(t, fmt.Errorf("status err order "))
				}

			} else {
				assert.NotEqual(t, err, nil)
			}

		})
	}
}

func TestOrderApplication_BuyCard(t *testing.T) {
	type args struct {
		amount            int64
		quantity          int64
		telco             string
		voucherCode       string
		userId            string
		sourceOfFund      string
		confirmOTPRequest order_system.ConfirmPaymentTokenRequest
	}

	th := NewTestOrderApplication()
	ctx := context_grpc.NewOrderSystemContextGRPC(context.TODO())
	defer th.DB.Drop(ctx)

	tests := []struct {
		name         string
		id           string
		expect       order_system.OrderStatus
		expectString string
		wantError    bool
		status       order_system.OrderStatus
		fd           func()
		args
	}{
		{
			args: args{
				amount:      1231232,
				quantity:    2,
				telco:       "VqweNP",
				voucherCode: "Voucherqwe",
				userId:      "user-qwe1",
			},
			name:         "FAIL(check quota fail)",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.MerchantFeeService.On("CheckMerchantQuotaAndFee", mock.Anything, mock.Anything).Return(
					&service_merchant_fee.CheckMerchantQuotaAndFeeRes{
						FeeAmount: 10000,
					}, nil)
				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{
					ErrorCode: "",
					Data:      []*entities2.ListLinked{},
				}, nil)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, errors.New(
					"han muc toi thieu la 500000d")).Times(1)

				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "VV",
						},
					},
				}, nil).Times(1)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "LL",
					},
				}, nil).Times(2)

				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{}, nil)
			},
		},
		{
			args: args{
				amount:      10000,
				quantity:    1,
				telco:       "VNP",
				voucherCode: "Voucher",
				userId:      "U",
			},
			name:         "happy_case",
			expect:       order_system.OrderStatus_ORDER_SUCCESS,
			expectString: "ORDER_SUCCESS",
			status:       order_system.OrderStatus_ORDER_SUCCESS,
			fd: func() {
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
				}, nil)

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "V",
						},
					},
				}, nil).Times(1)

				th.MerchantFeeService.On("GetMerchantVendorDiscount", ctx, mock.Anything).Return(&service_merchant_fee.GetMerchantVendorDiscountRes{
					MerchantDiscountAmount: 100,
					VendorDiscountAmount:   200,
				}, nil).Once()

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

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

				th.Card.On("BuyCard", ctx, mock.Anything).Return(&service_card.BuyCardRes{
					Cards: []*service_card.CardObjDTO{{
						Provider:   "VTM",
						CardNumber: "CardNumber",
						Serial:     "Serial",
						Price:      "",
					}},
				}, nil).Times(1)

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

			},
		},
		{
			args: args{
				amount:      10000,
				quantity:    1,
				telco:       "VNP",
				voucherCode: "Voucher",
				userId:      "user-1",
			},
			name:         "fail_case (buycard-fail)",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id: "user-1",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 100000,
							AmountFreeze:    0,
						}},
					},
				}, nil).Times(5)

				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "V",
						},
					},
				}, nil).Times(1)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

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

				th.Card.On("BuyCard", ctx, mock.Anything).Return(&service_card.BuyCardRes{
					Cards: []*service_card.CardObjDTO{{}},
				}, errors.New("some thing went wrong")).Once()

				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{}, nil)
			},
		},
		{
			args: args{
				amount:      10000,
				quantity:    1,
				telco:       "VNP",
				voucherCode: "Voucher",
				userId:      "user-1",
			},
			name:         "fail_case (payer_id not enough money)",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
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
				}, nil).Times(1)

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
				}, nil).Times(1)

				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "V",
						},
					},
				}, nil).Times(1)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, errors.New("Not enough money")).Times(1)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("UpdateTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{}, nil)
			},
		},
		{
			args: args{
				amount:      10000,
				quantity:    1,
				telco:       "VNP",
				voucherCode: "Voucher-invalid",
				userId:      "user-1",
			},
			name:         "fail_case (voucher is invalid)",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
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
				}, nil).Times(1)

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
				}, nil).Times(1)

				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "V",
						},
					},
				}, errors.New("[CheckUserCanUseVoucher] cannot find voucher")).Times(1)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

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
		{
			args: args{
				amount:      10000,
				quantity:    1,
				telco:       "VNP",
				voucherCode: "Voucher",
				userId:      "U1",
			},
			name:         "PENDING CASE (PENDING TRANSACTION SERVICE CARD)",
			expect:       order_system.OrderStatus_ORDER_VERIFYING,
			expectString: "ORDER_VERIFYING",
			status:       order_system.OrderStatus_ORDER_VERIFYING,
			fd: func() {
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id: "U1",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 100000,
							AmountFreeze:    0,
						}},
					},
				}, nil).Times(5)

				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "V",
						},
					},
				}, nil).Times(1)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("UpdateTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

				th.Card.On("BuyCard", ctx, mock.Anything).Return(&service_card.BuyCardRes{
					Status: "PENDING",
					Cards:  []*service_card.CardObjDTO{{}},
				}, nil).Once()

				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{}, nil)

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fd()
			res := &order_system.OrderBuyCardResponse{
				OrderEntity: nil,
				Cards:       nil,
			}

			order, err := th.OrderApplication.BuyCard(ctx, &order_system.OrderBuyCardRequest{
				Telco: tt.args.telco,
				OrderRequest: &order_system.OrderRequest{
					Amount:       tt.args.amount,
					Quantity:     tt.args.quantity,
					VoucherCode:  tt.args.voucherCode,
					UserID:       tt.args.userId,
					ServiceID:    constants.SUB_TRANSTYPE_WALLET_BUY_CARD,
					SubTransType: constants.SUB_TRANSTYPE_WALLET_BUY_CARD,
					SourceOfFund: tt.args.sourceOfFund,
				},
				ConfirmPaymentTokenRequest: &tt.args.confirmOTPRequest,
			}, res)

			if tt.wantError == false {
				if err != nil {
					assert.Error(t, err)
				}
				if !assert.Equal(t, order.Status.StatusOrderProto(), tt.expect) || !assert.Equal(t, order.Status.StatusString(), tt.expectString) {
					assert.Error(t, fmt.Errorf("status err order "))
				}

			} else {
				assert.NotEqual(t, err, nil)
			}
		})

	}

}

func TestOrderApplication_TopUp(t *testing.T) {
	th := NewTestOrderApplication()
	ctx := context.TODO()
	defer th.DB.Drop(ctx)
	tests := []struct {
		name              string
		id                string
		expect            order_system.OrderStatus
		expectString      string
		wantError         bool
		status            order_system.OrderStatus
		fd                func()
		confirmOTPRequest order_system.ConfirmPaymentTokenRequest
	}{
		{
			name:         "happy_case",
			expect:       order_system.OrderStatus_ORDER_SUCCESS,
			expectString: "ORDER_SUCCESS",
			status:       order_system.OrderStatus_ORDER_SUCCESS,
			fd: func() {
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "U1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 100000,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "V",
						},
					},
				}, nil).Times(1)

				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{
					ErrorCode: "",
					Data:      []*entities2.ListLinked{},
				}, nil)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

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
				}, nil).Once()

				th.Card.On("Topup", ctx, mock.Anything).Return(&service_card.TopupRes{}, nil).Times(1)
				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

			},
		},
		{
			name:         "fail-case(payer_not_enough_money)",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
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
				}, nil).Times(1)

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
				}, nil).Times(1)

				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "V",
						},
					},
				}, nil).Times(1)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, errors.New("not enough money")).Times(1)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("UpdateTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Once()

				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{}, nil)
			},
		},
		{
			name:         "fail-case(voucher is invalid)",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
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
				}, nil).Times(1)

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
				}, nil).Times(1)

				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "V",
						},
					},
				}, errors.New("[CheckUserCanUseVoucher] cannot find voucher")).Times(1)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

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
				}, nil).Once()

				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{},
					nil)
			},
		},
		{
			name:         "fail-case(top up service card failed)",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
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
				}, nil).Times(1)

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
				}, nil).Times(1)

				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "V",
						},
					},
				}, nil).Times(1)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

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
				}, nil).Once()

				th.Card.On("Topup", ctx, mock.Anything).Return(&service_card.TopupRes{}, errors.New("Unexpected Error")).Times(1)

				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{},
					nil)
			},
		},
		{
			name:         "PEDING CASE (top up service card TIMEOUT)",
			expect:       order_system.OrderStatus_ORDER_VERIFYING,
			expectString: "ORDER_VERIFYING",
			status:       order_system.OrderStatus_ORDER_VERIFYING,
			fd: func() {
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
				}, nil).Times(1)

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
				}, nil).Times(1)

				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "V",
						},
					},
				}, nil).Times(1)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

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
				}, nil).Once()

				th.Card.On("Topup", ctx, mock.Anything).Return(&service_card.TopupRes{
					Status: constants.TRANSACTION_STATUS_PENDING,
				}, nil).Times(1)

				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{},
					nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fd()
			res := &order_system.OrderTopUpResponse{
				OrderEntity: nil,
			}

			order, err := th.OrderApplication.TopUp(ctx, &order_system.OrderTopUpRequest{
				Telco: "VTM",
				OrderRequest: &order_system.OrderRequest{
					Amount:       10000,
					VoucherCode:  "V",
					UserID:       "U",
					ServiceID:    constants.SUB_TRANSTYPE_WALLET_TOPUP_CARD,
					SubTransType: constants.SUB_TRANSTYPE_WALLET_TOPUP_CARD,
					PhoneTopup:   "0912333123",
				},
				ConfirmPaymentTokenRequest: &tt.confirmOTPRequest,
			}, res)

			if tt.wantError == false {
				if err != nil {
					assert.Error(t, err)
				}

				if !assert.Equal(t, order.Status.StatusOrderProto(), tt.expect) || !assert.Equal(t, order.Status.StatusString(), tt.expectString) {
					assert.Error(t, fmt.Errorf("status err order "))
				}

			} else {
				assert.NotEqual(t, err, nil)
			}

		})
	}
}

func TestOrderApplication_PayBill(t *testing.T) {
	th := NewTestOrderApplication()
	ctx := context.TODO()
	defer th.DB.Drop(ctx)

	type args struct {
		serviceCode  string
		customerRef  string
		userId       string
		voucherCode  string
		amount       int64
		verifyOTP    order_system.ConfirmPaymentTokenRequest
		sourceOfFund string
	}

	tests := []struct {
		name         string
		id           string
		expect       order_system.OrderStatus
		expectString string
		wantError    bool
		status       order_system.OrderStatus
		fd           func()
		args
	}{
		{
			name: "fail_case_service_bill_failed",
			args: args{
				serviceCode: "505155",
				customerRef: "PCB0001233125",
				userId:      "user3",
				amount:      250000,
			},
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "U",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 100000,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{
					ErrorCode: "",
					Data:      []*entities2.ListLinked{},
				}, nil)

				th.Card.On("FindVendorByCode", ctx, &service_card.FindVendorByCodeReq{
					ServiceCode: "505155",
				}).Return(&service_card.FindVendorByCodeRes{
					Vendor: &service_card.VendorDTO{
						Code: "505155",
						Type: constants.SUB_TRANSTYPE_WALLET_PAY_BILL_ELECTRIC,
					},
				}, nil)
				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "V",
						},
					},
				}, nil).Times(1)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

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
				}, nil).Times(2)

				th.Card.On("PaidBill", ctx, mock.Anything).Return(&service_card.PaidBillRes{}, errors.New("Unexpected Error service card")).Once()

				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{},
					nil)

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

			},
		},
		{
			name: "PENDING",
			args: args{
				serviceCode: "505155",
				customerRef: "PCB0001233125",
				userId:      "user3",
				amount:      250000,
			},
			expect:       order_system.OrderStatus_ORDER_VERIFYING,
			expectString: "ORDER_VERIFYING",
			status:       order_system.OrderStatus_ORDER_VERIFYING,
			fd: func() {
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "U",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 100000,
							AmountFreeze:    0,
						}},
					},
				}, nil).Times(1)

				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "U",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 100000,
							AmountFreeze:    0,
						}},
					},
				}, nil).Times(1)

				th.Card.On("FindVendorByCode", ctx, &service_card.FindVendorByCodeReq{
					ServiceCode: "505155",
				}).Return(&service_card.FindVendorByCodeRes{
					Vendor: &service_card.VendorDTO{
						Code: "505155",
						Type: constants.SUB_TRANSTYPE_WALLET_PAY_BILL_ELECTRIC,
					},
				}, nil)
				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "V",
						},
					},
				}, nil).Times(1)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

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
				}, nil).Times(2)

				th.Card.On("PaidBill", ctx, mock.Anything).Return(&service_card.PaidBillRes{
					Status: constants.TRANSACTION_STATUS_PENDING,
				}, nil).Once()

				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{},
					nil)

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

			},
		},
		{
			args: args{
				serviceCode: "505153",
				customerRef: "PCB0001233123",
				userId:      "user1",
				amount:      250000,
			},
			name:         "happy_case",
			expect:       order_system.OrderStatus_ORDER_SUCCESS,
			expectString: "ORDER_SUCCESS",
			status:       order_system.OrderStatus_ORDER_SUCCESS,
			fd: func() {
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
				}, nil).Times(1)

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
				}, nil).Times(1)

				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "V",
						},
					},
				}, nil).Times(1)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

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

				th.Card.On("FindVendorByCode", ctx, &service_card.FindVendorByCodeReq{
					ServiceCode: "505153",
				}).Return(&service_card.FindVendorByCodeRes{
					Vendor: &service_card.VendorDTO{
						Code: "505153",
						Type: constants.SUB_TRANSTYPE_WALLET_PAY_BILL_ELECTRIC,
					},
				}, nil)

				th.Card.On("PaidBill", ctx, mock.Anything).Return(&service_card.PaidBillRes{}, nil)

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

			},
		},
		{
			args: args{
				serviceCode: "505154",
				customerRef: "PCB0001233124",
				userId:      "user1",
				amount:      250000,
			},
			name:         "fail_case(payer is not enough money)",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
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
				}, nil).Times(1)

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
				}, nil).Times(1)

				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "V",
						},
					},
				}, nil).Times(1)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, errors.New("Not enough money")).Times(1)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("UpdateTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(2)

				th.Card.On("FindVendorByCode", ctx, &service_card.FindVendorByCodeReq{
					ServiceCode:          "505154",
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     nil,
					XXX_sizecache:        0,
				}).Return(&service_card.FindVendorByCodeRes{
					Vendor: &service_card.VendorDTO{
						Code: "505154",
						Type: constants.SUB_TRANSTYPE_WALLET_PAY_BILL_ELECTRIC,
					},
				}, nil)

				th.Card.On("PaidBill", ctx, mock.Anything).Return(&service_card.PaidBillRes{}, nil)

				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{}, nil)

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fd()
			res := &order_system.OrderPayBillResponse{
				OrderEntity: nil,
			}

			order, err := th.OrderApplication.PayBill(ctx, &order_system.OrderPayBillRequest{
				ServiceCode: tt.args.serviceCode,
				BillingCode: tt.args.customerRef,
				OrderRequest: &order_system.OrderRequest{
					Amount:       tt.args.amount,
					VoucherCode:  tt.args.voucherCode,
					UserID:       tt.args.userId,
					ServiceID:    constants.SUB_TRANSTYPE_WALLET_PAY_BILL_LOAN,
					SubTransType: constants.SUB_TRANSTYPE_WALLET_PAY_BILL_LOAN,
					SourceOfFund: tt.args.sourceOfFund,
				},
			}, res)

			if tt.wantError == false {
				if err != nil {
					assert.Error(t, err)
				}

				if !assert.Equal(t, order.Status.StatusOrderProto(), tt.expect) || !assert.Equal(t, order.Status.StatusString(), tt.expectString) {
					assert.Error(t, fmt.Errorf("status err order "))
				}

			} else {
				assert.NotEqual(t, err, nil)
			}

		})
	}
}
func TestOrderApplication_FundWal2Bank(t *testing.T) {
	th := NewTestOrderApplication()
	ctx := context.TODO()
	defer th.DB.Drop(ctx)
	tests := []struct {
		name         string
		id           string
		expect       order_system.OrderStatus
		expectString string
		wantError    bool
		status       order_system.OrderStatus
		fd           func()
	}{
		{
			name:         "happy_case",
			expect:       order_system.OrderStatus_ORDER_SUCCESS,
			expectString: "ORDER_SUCCESS",
			status:       order_system.OrderStatus_ORDER_SUCCESS,
			fd: func() {
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
				}, nil)

				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "V",
						},
					},
				}, nil).Times(1)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Bank.On("GetBankByCode", mock.AnythingOfType("string")).Return(entities.Bank{
					BankId:    1,
					Name:      "bank",
					ShortName: "bank1",
					IBFTCode:  "",
				}, nil).Once()

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.MerchantFeeService.On("CheckMerchantQuotaAndFee", ctx, mock.Anything).Return(&service_merchant_fee.CheckMerchantQuotaAndFeeRes{
					FeeAmount: 10,
					FeeMethod: 0,
				}, nil).Once()

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
				}, nil).Once()

				th.BankService.On("IBFTInquiry", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(entities2.IBFTInquiryCheckResponse{}, nil).Times(1)
				th.BankService.On("IBFTTransfer", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
					mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("int64"), mock.AnythingOfType("string")).Return(entities2.IBFTTransferResponse{}, nil).Once()

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

			},
		},
		{
			name:         "verifying case",
			expect:       order_system.OrderStatus_ORDER_VERIFYING,
			expectString: "ORDER_VERIFYING",
			status:       order_system.OrderStatus_ORDER_VERIFYING,
			fd: func() {
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
				}, nil)

				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "V",
						},
					},
				}, nil).Times(1)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Bank.On("GetBankByCode", mock.AnythingOfType("string")).Return(entities.Bank{
					BankId:    1,
					Name:      "bank",
					ShortName: "bank1",
					IBFTCode:  "",
				}, nil).Once()

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.MerchantFeeService.On("CheckMerchantQuotaAndFee", ctx, mock.Anything).Return(&service_merchant_fee.CheckMerchantQuotaAndFeeRes{
					FeeAmount: 10,
					FeeMethod: 0,
				}, nil).Once()

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
				}, nil).Once()

				th.BankService.On("IBFTTransfer", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
					mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("int64"), mock.AnythingOfType("string")).Return(entities2.IBFTTransferResponse{
					ErrorCode: "102",
				}, nil).Once()

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fd()
			res := &order_system.FundWallet2BankResponse{
				OrderEntity: nil,
			}

			order, err := th.OrderApplication.FundWal2Bank(ctx, &order_system.FundWallet2BankRequest{
				AccountNo:   "",
				CardNo:      "card",
				BankCode:    "",
				Description: "gui tien merchant -> user",
				OrderRequest: &order_system.OrderRequest{
					Amount:       20000,
					MerchantID:   "merchant1",
					ServiceID:    constants.TRANSTYPE_WALLET_TRANS2BANK,
					TransType:    constants.TRANSTYPE_BEHALF_PAY_WALLET_TO_BANK,
					SourceOfFund: constants.SOURCE_OF_FUND_BALANCE_WALLET,
					DeviceID:     "D",
					RefID:        "merchant-ref-id",
				},
			}, res)

			if tt.wantError == false {
				if err != nil {
					assert.Error(t, err)
				}
				if !assert.Equal(t, tt.expect, order.Status.StatusOrderProto()) || !assert.Equal(t, tt.expectString, order.Status.StatusString()) {
					assert.Error(t, fmt.Errorf("status err order "))
				}

			} else {
				assert.NotEqual(t, err, nil)
			}

		})
	}
}
func TestOrderApplication_FundWal2Wal(t *testing.T) {
	th := NewTestOrderApplication()
	ctx := context.TODO()
	defer th.DB.Drop(ctx)
	tests := []struct {
		name         string
		id           string
		expect       order_system.OrderStatus
		expectString string
		wantError    bool
		status       order_system.OrderStatus
		fd           func()
	}{
		{
			name:         "happy_case",
			expect:       order_system.OrderStatus_ORDER_SUCCESS,
			expectString: "ORDER_SUCCESS",
			status:       order_system.OrderStatus_ORDER_SUCCESS,
			fd: func() {
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
				}, nil)

				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "V",
						},
					},
				}, nil).Times(1)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.User.On("FindUserByPhone", ctx, mock.Anything).Return(&service_user.FindUserByPhoneResponse{
					User: &service_user.UserDTO{Id: "payee id", HasEverLinkedBank: true},
				}, nil).Once()

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.MerchantFeeService.On("CheckMerchantQuotaAndFee", ctx, mock.Anything).Return(&service_merchant_fee.CheckMerchantQuotaAndFeeRes{
					FeeAmount: 100,
					FeeMethod: 0,
				}, nil).Once()

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
				}, nil).Once()

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

			},
		},
		{
			name:         "fail_case(unexpected init transaction)",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
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
				}, nil).Times(1)

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
				}, nil).Times(1)

				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "V",
						},
					},
				}, nil).Times(1)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.MerchantFeeService.On("CheckMerchantQuotaAndFee", ctx, mock.Anything).Return(&service_merchant_fee.CheckMerchantQuotaAndFeeRes{
					FeeAmount: 10000,
					FeeMethod: 0,
				}, nil).Once()

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, errors.New("payee is not exist")).Times(1)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("UpdateTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Once()

				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{}, nil)
				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fd()
			res := &order_system.FundWallet2WalletResponse{
				OrderEntity: nil,
			}

			order, err := th.OrderApplication.FundWal2Wal(ctx, &order_system.FundWallet2WalletRequest{
				ToPhoneNumber: "0912223123",
				Description:   "gui tien merchant -> user ko ton tai",
				OrderRequest: &order_system.OrderRequest{
					Amount:       20000,
					MerchantID:   "merchant1",
					ServiceID:    constants.SERVICE_TYPE_COLLECTION_AND_PAY,
					TransType:    constants.TRANSTYPE_BEHALF_PAY_WALLET_TO_WALLET,
					SourceOfFund: constants.SOURCE_OF_FUND_BALANCE_WALLET,
					DeviceID:     "D",
					RefID:        "merchant-ref-id",
				},
			}, res)

			if tt.wantError == false {
				if err != nil {
					assert.Error(t, err)
				}
				if !assert.Equal(t, order.Status.StatusOrderProto(), tt.expect) || !assert.Equal(t, order.Status.StatusString(), tt.expectString) {
					assert.Error(t, fmt.Errorf("status err order "))
				}

			} else {
				assert.NotEqual(t, err, nil)
			}

		})
	}
}

func TestOrderApplication_CashIn(t *testing.T) {
	th := NewTestOrderApplication()
	ctx := context.TODO()
	defer th.DB.Drop(ctx)
	type args struct {
		amount int64
		userId string
	}

	tests := []struct {
		name         string
		id           string
		expect       order_system.OrderStatus
		expectString string
		wantError    bool
		status       order_system.OrderStatus
		fd           func()
		args
	}{
		////{
		////	name:         "happy_case",
		////	expect:       order_system.OrderStatus_ORDER_PROCESSING,
		////	expectString: "ORDER_PROCESSING",
		////	status:       order_system.OrderStatus_ORDER_PROCESSING,
		////	fd: func() {
		////		th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
		////			TransactionId: "T",
		////		}, nil).Times(1)
		////
		////		th.BankService.On("LinkInfo",mock.AnythingOfType("int64")).Return(entities2.LinkInfoResponse{
		////			ErrorCode: "000",
		////			Data:      entities2.DataLinkInfoResponse{
		////				IBFTCode: "GPSTB",
		////			},
		////		},nil)
		////
		////		th.BankService.On("CashIn",mock.Anything,mock.Anything,mock.Anything,mock.Anything,mock.Anything).Return(entities2.CashInResponse{
		////			ErrorCode: "000",
		////			Data:      entities2.CashInDataResponse{
		////				BankTraceID:      "1111",
		////				RefDataVerifyOTP: nil,
		////				LinkID:           1111,
		////			},
		////		},nil)
		////	},
		////	args : args{
		////		amount: 11011,
		////		userId: "xxx",
		////	},
		////},
		{
			name:         "FAIL CASE (service_bankgw_fail)",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id: "user-qwe1",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(service_transaction.ETransactionDTO{
					Id: "T",
				}, nil)

				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{}, nil)
				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(nil, nil)

				th.BankService.On("LinkInfo", mock.AnythingOfType("int64")).Return(entities2.LinkInfoResponse{
					ErrorCode: "000",
					Data: entities2.DataLinkInfoResponse{
						BankCode:   "GPBIDV",
						GpayUserID: "11111",
					},
				}, nil)

				th.BankService.On("CashIn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities2.CashInResponse{}, errors.New("UserID khong hop le"))

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T123",
				}, nil).Times(1)

				th.Transaction.On("UpdateTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T123",
				}, nil)

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T123",
				}, nil)
				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

			},
			args: args{
				amount: 11111,
				userId: "111112",
			},
		},
		{
			name:         "PENDING CASE (service_bank_gw_pending)",
			expect:       order_system.OrderStatus_ORDER_VERIFYING,
			expectString: "ORDER_VERIFYING",
			status:       order_system.OrderStatus_ORDER_VERIFYING,
			fd: func() {
				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T123",
				}, nil).Times(1)

				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id: "2",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 100000,
							AmountFreeze:    0,
						}},
					},
				}, nil).Times(1)
				//
				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{
					FeeAmount: 1,
				}, nil).Times(1)

				th.BankService.On("LinkInfo", mock.AnythingOfType("int64")).Return(entities2.LinkInfoResponse{
					ErrorCode: "000",
					Data: entities2.DataLinkInfoResponse{
						GpayUserID: "2",
					},
				}, nil)

				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{
					ErrorCode: "",
					Data:      []*entities2.ListLinked{},
				}, nil)

				th.BankService.On("CashIn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities2.CashInResponse{
					ErrorCode: "306",
					Message:   "TIME OUT",
				}, nil)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T123",
				}, nil).Times(1)

				th.Transaction.On("UpdateTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T123",
				}, nil)

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T123",
				}, nil)
			},
			args: args{
				amount: 22222,
				userId: "2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fd()
			res := &order_system.OrderCashInResponse{}

			order, err := th.OrderApplication.CashIn(ctx, &order_system.OrderCashInRequest{
				Description: "",
				OrderRequest: &order_system.OrderRequest{
					Amount:       tt.args.amount,
					ServiceID:    constants.TRANSTYPE_WALLET_CASH_IN,
					TransType:    constants.TRANSTYPE_WALLET_CASH_IN,
					SourceOfFund: constants.SOURCE_OF_FUND_BALANCE_WALLET,
					ToUserID:     tt.args.userId,
				},
			}, res)

			if tt.wantError == false {
				if err != nil {
					assert.Error(t, err)
				}

				if order.Status.StatusOrderProto() == order_system.OrderStatus_UNKNOWN {
					assert.Error(t, errors.New("Invalid link id"))
				}

			} else {
				if !assert.Equal(t, order.Status.StatusOrderProto(), tt.expect) || !assert.Equal(t, order.Status.StatusString(), tt.expectString) {
					assert.Error(t, fmt.Errorf("status err order "))
				}
			}

		})
	}
}

func TestOrderApplication_VerifyOTP(t *testing.T) {
	type args struct {
		orderID string
		otp     string
		userId  string
	}
	th := NewTestOrderApplication()
	ctx := context.TODO()
	defer th.DB.Drop(ctx)

	orderInit, _ := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		ServiceID:    "123",
		UserID:       "123",
		ServiceType:  "123",
		OrderType:    "123",
		SubOrderType: "23",
		Amount:       222220,
		SourceOfFund: "xczxc",
		Status:       2,
		VoucherCode:  "asd",
		DeviceID:     "asd",
		BankCode:     "STB",
		PhoneTopUp:   "",
		ExpiredAt:    helpers.GetCurrentTime().Add(10 * time.Hour),
	})
	orderInitSuccesOrder, _ := th.OrderApplication.SuccessOrder(ctx, orderInit)

	orderInit2, _ := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		ServiceID:    "123",
		UserID:       "2",
		ServiceType:  "123",
		OrderType:    "123",
		SubOrderType: "23",
		Amount:       222220,
		SourceOfFund: "xczxc",
		Status:       2,
		VoucherCode:  "asd",
		DeviceID:     "asd",
		BankCode:     "STB",
		PhoneTopUp:   "",
	})
	orderInitProccessing, _ := th.OrderApplication.ProcessingOrder(ctx, orderInit2)

	orderInit3, _ := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		ServiceID:    "123",
		UserID:       "2",
		ServiceType:  "123",
		OrderType:    "123",
		SubOrderType: "23",
		Amount:       222220,
		SourceOfFund: "xczxc",
		Status:       2,
		VoucherCode:  "asd",
		DeviceID:     "asd",
		BankCode:     "STB",
		PhoneTopUp:   "",
	})
	orderInitProccessing3, _ := th.OrderApplication.ProcessingOrder(ctx, orderInit3)

	tests := []struct {
		name         string
		id           string
		expect       order_system.OrderStatus
		expectString string
		wantError    bool
		status       order_system.OrderStatus
		fd           func()
		args
	}{
		//{
		//	args: args{
		//		orderID: orderInit.OrderID,
		//		otp:     "1234561",
		//	},
		//	name:         "happy case",
		//	expect:       order_system.OrderStatus_ORDER_SUCCESS,
		//	expectString: "ORDER_SUCCESS",
		//	status:       order_system.OrderStatus_ORDER_SUCCESS,
		//	fd: func() {
		//		th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
		//			TransactionId: "T",
		//		}, nil).Times(1)
		//
		//		th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)
		//
		//		th.BankService.On("VerifyOTP", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities2.VerifyOTPResponse{}, nil)
		//
		//		th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
		//			UserDetail: &service_user.UserDetailDTO{
		//				User: &service_user.UserDTO{
		//					Id: "U",
		//				},
		//				Balances: []*service_user.BalanceDTO{{
		//					AmountAvailable: 100000,
		//					AmountFreeze:    0,
		//				}},
		//			},
		//		}, nil).Times(1)
		//
		//		th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
		//			{
		//				ID: "L",
		//			},
		//		}, nil).Times(2)
		//
		//		th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
		//			TransactionId: "T",
		//		}, nil).Times(1)
		//
		//		th.Transaction.On("UpdateTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
		//			TransactionId: "T",
		//		}, nil)
		//
		//		th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
		//			TransactionId: "T",
		//		}, nil)
		//	},
		//},
		{
			args: args{
				orderID: orderInitProccessing.OrderID,
				otp:     "1234561",
				userId:  "2",
			},
			name:         "fail_case(bank gw fail)",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.TransactionDTO{}, nil).Times(1)

				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: orderInit2.TransactionID,
				}, nil)
				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil)
				th.BankService.On("LinkInfo", mock.Anything).Return(entities2.LinkInfoResponse{
					Data: entities2.DataLinkInfoResponse{
						GpayUserID: "2",
					},
				}, nil)

				th.BankService.On("VerifyOTP", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities2.VerifyOTPResponse{}, errors.New("bank gw err unexpected")).Times(1)

				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id: "2",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 100000,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.OrderRepository.On("ReplaceByID", ctx, mock.Anything).Return(&entities.OrderEntity{
					Status: entities.EntityStatus(5),
				}, nil)

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
		{
			args: args{
				orderID: orderInitSuccesOrder.OrderID,
				otp:     "12345617",
				userId:  "1",
			},
			name:         "invalid orderid",
			expect:       order_system.OrderStatus_ORDER_PROCESSING,
			expectString: "ORDER_PROCESSING",
			status:       order_system.OrderStatus_ORDER_PROCESSING,
			fd: func() {
				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: orderInit2.TransactionID,
				}, nil)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.BankService.On("LinkInfo", mock.Anything).Return(entities2.LinkInfoResponse{
					Data: entities2.DataLinkInfoResponse{
						GpayUserID: "123",
					},
				}, nil)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id: "1",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 100000,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.OrderRepository.On("ReplaceByID", ctx, mock.Anything).Return(&entities.OrderEntity{}, nil)

			},
			wantError: true,
		},
		{
			args: args{
				orderID: orderInitProccessing3.OrderID,
				otp:     "1234561",
				userId:  "2",
			},
			name:         "fail_case (max wrong OTP fail)",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.TransactionDTO{}, nil).Times(1)

				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: orderInit2.TransactionID,
				}, nil)
				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil)
				th.BankService.On("LinkInfo", mock.Anything).Return(entities2.LinkInfoResponse{
					Data: entities2.DataLinkInfoResponse{
						GpayUserID: "2",
					},
				}, nil)

				th.BankService.On("VerifyOTP", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities2.VerifyOTPResponse{
					ErrorCode: "461",
				}, errors.New("Max attempt wrong OTP")).Times(1)

				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id: "2",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 100000,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.OrderRepository.On("ReplaceByID", ctx, mock.Anything).Return(&entities.OrderEntity{
					Status: entities.EntityStatus(5),
				}, nil)

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
			res := &order_system.OrderCashOTPResponse{
				OrderEntity: nil,
			}

			order, _ := th.OrderApplication.VerifyOTP(ctx, &order_system.OrderCashOTPRequest{
				RefBankTraceID: "11",
				LinkId:         1,
				Otp:            tt.args.otp,
				OrderId:        tt.args.orderID,
				UserId:         tt.args.userId,
			}, res)

			if tt.wantError == false {
				if !assert.Equal(t, tt.expect, order.Status.StatusOrderProto()) {
					assert.Error(t, errors.New("compare status order failed"))
				}

			}
		})

	}

}

func TestOrderApplication_CashOut(t *testing.T) {
	th := NewTestOrderApplication()
	ctx := context.TODO()
	defer th.DB.Drop(ctx)
	type args struct {
		amount int64
		userId string
	}

	tests := []struct {
		name         string
		id           string
		expect       order_system.OrderStatus
		expectString string
		wantError    bool
		status       order_system.OrderStatus
		fd           func()
		args
	}{
		//{
		//	name:         "fail_case(service_bank_gw_FAIL)",
		//	expect:       order_system.OrderStatus_ORDER_FAILED,
		//	expectString: "ORDER_FAILED",
		//	status:       order_system.OrderStatus_ORDER_FAILED,
		//	fd: func() {
		//		th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
		//			TransactionId: "T123",
		//		}, nil).Times(1)
		//
		//		th.BankService.On("LinkInfo", mock.AnythingOfType("int64")).Return(entities2.LinkInfoResponse{
		//			ErrorCode: "000",
		//			Data: entities2.DataLinkInfoResponse{
		//				IBFTCode: "GPBIDV",
		//			},
		//		}, nil)
		//
		//		th.BankService.On("CashOut", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,mock.Anything).Return(entities2.CashOutResponse{}, errors.New("Unexpected Error BankGW"))
		//		th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)
		//
		//		th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
		//			TransactionId: "T123",
		//		}, nil).Times(1)
		//
		//		th.Transaction.On("UpdateTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
		//			TransactionId: "T123",
		//		}, nil)
		//
		//		th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
		//			TransactionId: "T123",
		//		}, nil)
		//		th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything ).Return(nil)
		//
		//	},
		//	args: args{
		//		amount: 22222,
		//		userId: "usssssserr",
		//	},
		//},
		//{
		//	name:         "pending case (service_bank_gw_time_out)",
		//	expect:       order_system.OrderStatus_ORDER_VERIFYING,
		//	expectString: "ORDER_VERIFYING",
		//	status:       order_system.OrderStatus_ORDER_VERIFYING,
		//	fd: func() {
		//		th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
		//			TransactionId: "T123",
		//		}, nil).Times(1)
		//
		//		th.BankService.On("LinkInfo", mock.AnythingOfType("int64")).Return(entities2.LinkInfoResponse{
		//			ErrorCode: "000",
		//			Data: entities2.DataLinkInfoResponse{
		//				IBFTCode: "GPBIDV",
		//			},
		//		}, nil)
		//
		//		th.BankService.On("CashOut", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,mock.Anything).Return(entities2.CashOutResponse{
		//			ErrorCode: "502",
		//			Message:   "PENDING TRANSACTION",
		//		}, nil)
		//		th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)
		//
		//		th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
		//			TransactionId: "T123",
		//		}, nil).Times(1)
		//
		//		th.Transaction.On("UpdateTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
		//			TransactionId: "T123",
		//		}, nil)
		//
		//		th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
		//			TransactionId: "T123",
		//		}, nil)
		//th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything ).Return(nil)

		//	},
		//	args: args{
		//		amount: 22222,
		//		userId: "usssssserr",
		//	},
		//},
		{
			name:         "FAIL case (invalid user id)",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T123",
				}, nil).Times(1)

				th.BankService.On("LinkInfo", mock.AnythingOfType("int64")).Return(entities2.LinkInfoResponse{
					ErrorCode: "000",
					Data: entities2.DataLinkInfoResponse{
						BankCode: "GPBIDV",
					},
				}, nil)

				th.BankService.On("CashOut", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities2.CashOutResponse{}, errors.New("User Id khong hop le"))
				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T123",
				}, nil).Times(1)

				th.Transaction.On("UpdateTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T123",
				}, nil)

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T123",
				}, nil)

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
				}, nil)

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

			},
			args: args{
				amount: 22222,
				userId: "usssssserr",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fd()
			res := &order_system.CashOutResponse{}

			order, err := th.OrderApplication.CashOut(ctx, &order_system.CashOutRequest{
				LinkId:      1,
				Description: "asd",
				OrderRequest: &order_system.OrderRequest{
					Amount: tt.args.amount,
				},
			}, res)

			if tt.wantError == false {
				if err != nil {
					assert.Error(t, err)
				}
				if !assert.Equal(t, order.Status.StatusOrderProto(), tt.expect) || !assert.Equal(t, order.Status.StatusString(), tt.expectString) {
					assert.Error(t, fmt.Errorf("status err order "))
				}

			} else {
				assert.NotEqual(t, err, nil)
			}

		})
	}
}

func TestOrderApplication_IBFTTransfer(t *testing.T) {
	th := NewTestOrderApplication()
	ctx := context.TODO()
	defer th.DB.Drop(ctx)
	tests := []struct {
		name         string
		id           string
		expect       order_system.OrderStatus
		expectString string
		wantError    bool
		status       order_system.OrderStatus
		fd           func()
	}{
		{
			name:         "happy_case",
			expect:       order_system.OrderStatus_ORDER_SUCCESS,
			expectString: "ORDER_SUCCESS",
			status:       order_system.OrderStatus_ORDER_SUCCESS,
			fd: func() {
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
				}, nil)

				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "V",
						},
					},
				}, nil).Times(1)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Bank.On("GetBankByCode", mock.AnythingOfType("string")).Return(entities.Bank{
					BankId:    1,
					Name:      "bank",
					ShortName: "bank1",
					IBFTCode:  "",
				}, nil).Once()

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.MerchantFeeService.On("CheckMerchantQuotaAndFee", ctx, mock.Anything).Return(&service_merchant_fee.CheckMerchantQuotaAndFeeRes{
					FeeAmount: 10000,
					FeeMethod: 0,
				}, nil).Once()

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
				}, nil).Once()

				th.BankService.On("IBFTInquiry", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(entities2.IBFTInquiryCheckResponse{}, nil).Times(1)
				th.BankService.On("IBFTTransfer", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
					mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("int64"), mock.AnythingOfType("string")).Return(entities2.IBFTTransferResponse{}, nil).Once()

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)
			},
		},
		{
			name:         "ibft_server_FAIL ",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
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
				}, nil)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Bank.On("GetBankByCode", mock.AnythingOfType("string")).Return(entities.Bank{
					BankId:    1,
					Name:      "bank",
					ShortName: "bank1",
					IBFTCode:  "",
				}, nil).Once()

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.MerchantFeeService.On("CheckMerchantQuotaAndFee", ctx, mock.Anything).Return(&service_merchant_fee.CheckMerchantQuotaAndFeeRes{
					FeeAmount: 10000,
					FeeMethod: 0,
				}, nil).Once()

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
				}, nil).Once()

				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{}, nil)

				th.BankService.On("IBFTInquiry", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(entities2.IBFTInquiryCheckResponse{}, nil).Times(1)
				th.BankService.On("IBFTTransfer", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
					mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("int64"), mock.AnythingOfType("string")).Return(entities2.IBFTTransferResponse{}, errors.New(
					"Loi giao dich ibft ")).Once()
			},
		},
		{
			name:         "PENDING CASE TRANSACTION => VERIFYING",
			expect:       order_system.OrderStatus_ORDER_VERIFYING,
			expectString: "ORDER_VERIFYING",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
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
				}, nil)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Bank.On("GetBankByCode", mock.AnythingOfType("string")).Return(entities.Bank{
					BankId:    1,
					Name:      "bank",
					ShortName: "bank1",
					IBFTCode:  "",
				}, nil).Once()

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.MerchantFeeService.On("CheckMerchantQuotaAndFee", ctx, mock.Anything).Return(&service_merchant_fee.CheckMerchantQuotaAndFeeRes{
					FeeAmount: 10000,
					FeeMethod: 0,
				}, nil).Once()

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
				}, nil).Once()

				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{}, nil)

				th.BankService.On("IBFTInquiry", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(entities2.IBFTInquiryCheckResponse{}, nil).Times(1)
				th.BankService.On("IBFTTransfer", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
					mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("int64"), mock.AnythingOfType("string")).Return(entities2.IBFTTransferResponse{
					ErrorCode: "102",
				}, nil).Once()
			},
		},
		{
			name:         "Not enough money",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
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
				}, nil)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Bank.On("GetBankByCode", mock.AnythingOfType("string")).Return(entities.Bank{
					BankId:    1,
					Name:      "bank",
					ShortName: "bank1",
					IBFTCode:  "",
				}, nil).Once()

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.MerchantFeeService.On("CheckMerchantQuotaAndFee", ctx, mock.Anything).Return(&service_merchant_fee.CheckMerchantQuotaAndFeeRes{
					FeeAmount: 10000,
					FeeMethod: 0,
				}, nil).Once()

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, errors.New("NOT ENOGH MONEY")).Times(1)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("UpdateTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Once()

				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{}, nil)
			},
		},
		{
			name:         "conditions check quota ",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
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
				}, nil)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Bank.On("GetBankByCode", mock.AnythingOfType("string")).Return(entities.Bank{
					BankId:    1,
					Name:      "bank",
					ShortName: "bank1",
					IBFTCode:  "",
				}, nil).Once()

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(
					&service_transaction.CheckTransactionQuotaAndFeeRes{}, errors.New("Vuot qua han muc cho phep"))

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, errors.New("Vuot qua han muc cho phep")).Times(1)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("UpdateTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Once()

				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fd()
			res := &order_system.IBFTTransferResponse{
				OrderEntity: nil,
			}

			order, err := th.OrderApplication.IBFTransfer(ctx, &order_system.IBFTTransferRequest{
				AccountNo:   "",
				IbftCode:    "",
				CardNo:      "card",
				Description: "gui tien merchant -> user",
				OrderRequest: &order_system.OrderRequest{
					Amount:       20000,
					MerchantID:   "merchant1",
					ServiceID:    constants.TRANSTYPE_WALLET_TRANS2BANK,
					TransType:    constants.TRANSTYPE_BEHALF_PAY_WALLET_TO_BANK,
					SourceOfFund: constants.SOURCE_OF_FUND_BALANCE_WALLET,
					DeviceID:     "D",
					RefID:        "merchant-ref-id",
				},
			}, res)

			if tt.wantError == false {
				if err != nil {
					assert.Error(t, err)
				}
				if !assert.Equal(t, order.Status.StatusOrderProto(), tt.expect) || !assert.Equal(t, order.Status.StatusString(), tt.expectString) {
					assert.Error(t, fmt.Errorf("status err order "))
				}

			} else {
				assert.NotEqual(t, err, nil)
			}

		})
	}
}

func TestOrderApplication_PaymentMerchant(t *testing.T) {
	th := NewTestOrderApplication()
	ctx := context.TODO()
	defer th.DB.Drop(ctx)
	type args struct {
		amount       int64
		userId       string
		subTransType string
		transType    string
		merchantId   string
		voucherCode  string
		serviceId    string
	}

	tests := []struct {
		name         string
		id           string
		expect       order_system.OrderStatus
		expectString string
		wantError    bool
		status       order_system.OrderStatus
		fd           func()
		args
	}{
		{
			name:         "happy_case",
			expect:       order_system.OrderStatus_ORDER_PROCESSING,
			expectString: "ORDER_PROCESSING",
			status:       order_system.OrderStatus_ORDER_PROCESSING,
			fd: func() {
				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.BankService.On("LinkInfo", mock.AnythingOfType("int64")).Return(entities2.LinkInfoResponse{
					ErrorCode: "000",
				}, nil).Once()

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.MerchantFeeService.On("CheckMerchantQuotaAndFee", mock.Anything, mock.Anything).Return(
					&service_merchant_fee.CheckMerchantQuotaAndFeeRes{
						FeeAmount: 10000,
					}, nil).Once()

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T123",
				}, nil).Times(1)

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{}, nil).Once()

				th.User.On("FindUserDetailById", mock.Anything, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
							Id:                "xxx",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 100000,
							AmountFreeze:    0,
						}},
					},
				}, nil).Times(8)

			},
			args: args{
				amount:       11011,
				userId:       "xxx",
				subTransType: "PAY_MERCHANT",
				transType:    "PAY2MERCHANT",
				merchantId:   "merchant1",
				voucherCode:  "",
				serviceId:    "WALLET",
			},
		},
		{

			name:         "fail_case (user thanh toan khong du tien)",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T1",
				}, errors.New("tai khoan khong du")).Times(1)

				th.BankService.On("LinkInfo", mock.AnythingOfType("int64")).Return(entities2.LinkInfoResponse{
					ErrorCode: "000",
				}, nil).Once()

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)
				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)
				th.MerchantFeeService.On("CheckMerchantQuotaAndFee", mock.Anything, mock.Anything).Return(
					&service_merchant_fee.CheckMerchantQuotaAndFeeRes{
						FeeAmount: 10000,
					}, nil).Once()

				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "VV",
						},
					},
				}, nil).Times(1)

				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{}, nil)

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{}, nil).Once()

				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "xxx",
							Kyc:               "ACTIVE",
							HasEverLinkedBank: true,
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 100000,
							AmountFreeze:    0,
						}},
					},
				}, nil)

			},
			args: args{
				amount:       11011,
				userId:       "xxx",
				subTransType: "PAY_MERCHANT",
				transType:    "PAY2MERCHANT",
				merchantId:   "merchant1",
				voucherCode:  "voucher1",
				serviceId:    "WALLET",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fd()
			res := &order_system.PaymentMerchantResponse{}

			order, err := th.OrderApplication.StaticQR(ctx, &order_system.PaymentMerchantRequest{
				OrderRequest: &order_system.OrderRequest{
					Amount:       tt.args.amount,
					VoucherCode:  tt.args.voucherCode,
					UserID:       tt.args.userId,
					MerchantID:   tt.args.merchantId,
					SubTransType: tt.args.subTransType,
					ServiceID:    tt.args.serviceId,
					TransType:    tt.args.transType,
					SourceOfFund: constants.SOURCE_OF_FUND_BALANCE_WALLET,
				},
			}, res)

			if tt.wantError == false {
				if err != nil {
					assert.Error(t, err)
				}
			} else {
				if !assert.Equal(t, order.Status.StatusOrderProto(), tt.expect) || !assert.Equal(t, order.Status.StatusString(), tt.expectString) {
					assert.Error(t, fmt.Errorf("status err order "))
				}
			}

		})
	}
}

func TestOrderApplication_ConfirmQrDynamicOrder(t *testing.T) {
	th := NewTestOrderApplication()
	ctx := context.TODO()
	defer th.DB.Drop(ctx)

	init_case1, err := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		ServiceID:           "",
		UserID:              "namnam",
		SubscribeMerchantID: "",
		TransactionID:       "",
		ServiceType:         "",
		OrderType:           "",
		SubOrderType:        "",
		Amount:              99999,
		SourceOfFund:        "",
	})

	init_case2, err := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		ServiceID:           "",
		UserID:              "namnam",
		SubscribeMerchantID: "",
		TransactionID:       "",
		ServiceType:         "",
		OrderType:           "",
		SubOrderType:        "",
		Amount:              99999,
		SourceOfFund:        "",
	})

	init_case3, err := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		ServiceID:           "",
		UserID:              "namnam",
		SubscribeMerchantID: "",
		TransactionID:       "",
		ServiceType:         "",
		OrderType:           "",
		SubOrderType:        constants.SUB_TRANSTYPE_WALLET_QR_DYNAMIC,
		Amount:              99999,
		SourceOfFund:        "",
	})

	if err != nil || init_case1 == nil || init_case2 == nil {
		t.Error("can not init order")
	}

	type args struct {
		orderId      string
		amount       int64
		voucherCode  string
		subTransType string
	}

	tests := []struct {
		name         string
		id           string
		expect       order_system.OrderStatus
		expectString string
		wantError    bool
		status       order_system.OrderStatus
		mockFunc     func()
		args
	}{
		{
			name:         "case 1 happy case",
			id:           "",
			expect:       order_system.OrderStatus_ORDER_SUCCESS,
			expectString: "ORDER_SUCCESS",
			wantError:    false,
			status:       order_system.OrderStatus_ORDER_SUCCESS,
			mockFunc: func() {
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)
				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{
					ErrorCode: "",
					Data:      []*entities2.ListLinked{},
				}, nil)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

			},
			args: args{
				orderId: init_case1.OrderID,
				amount:  1111,
			},
		},
		{
			name:         "case 2 fail case (tai khoan khong du)",
			id:           "",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			wantError:    true,
			status:       order_system.OrderStatus_ORDER_FAILED,
			mockFunc: func() {
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)
				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{
					ErrorCode: "",
					Data:      []*entities2.ListLinked{},
				}, nil)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, errors.New("Not enough money")).Times(1)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{}, nil)

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

			},
			args: args{
				orderId: init_case2.OrderID,
				amount:  2233,
			},
		},
		{
			name:         "case 3 Sub Trans Type WEBTOAPP",
			id:           "",
			expect:       order_system.OrderStatus_ORDER_SUCCESS,
			expectString: "ORDER_SUCCESS",
			wantError:    false,
			status:       order_system.OrderStatus_ORDER_SUCCESS,
			mockFunc: func() {
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)
				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{
					ErrorCode: "",
					Data:      []*entities2.ListLinked{},
				}, nil)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

			},
			args: args{
				orderId:      init_case3.OrderID,
				amount:       1111,
				subTransType: "WEBTOAPP",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			res := &order_system.ConfirmOrderResponse{}

			order, err := th.OrderApplication.ConfirmOrder(ctx, &order_system.ConfirmOrderRequest{
				OrderId: tt.orderId,
				OrderRequest: &order_system.OrderRequest{
					Amount:       tt.amount,
					VoucherCode:  tt.voucherCode,
					SubTransType: tt.subTransType,
				}}, res)

			if err != nil {
				t.Log(err)
			}
			if !assert.Equal(t, order.Status.StatusOrderProto(), tt.expect) || !assert.Equal(t, order.Status.StatusString(), tt.expectString) {
				assert.Error(t, fmt.Errorf("status err order "))
			} else {
				if tt.subTransType == constants.SUB_TRANSTYPE_WALLET_WEB_TO_APP {
					assert.Equal(t, res.OrderEntity.SubOrderType, tt.subTransType)
				}
			}

		})
	}

}

func TestOrderApplication_UpdateCreditPayment(t *testing.T) {
	th := NewTestOrderApplication()
	ctx := context.TODO()
	defer th.DB.Drop(ctx)

	init_case1, err := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		ServiceID:           "",
		SubscribeMerchantID: "",
		TransactionID:       "T",
		ServiceType:         "",
		OrderType:           "",
		SubOrderType:        "",
		Amount:              99999,
		SourceOfFund:        "",
	})

	init_case2, err := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		ServiceID:           "",
		SubscribeMerchantID: "",
		TransactionID:       "T1",
		ServiceType:         "",
		OrderType:           "",
		SubOrderType:        "",
		Amount:              99999,
		SourceOfFund:        "",
	})

	init_case_verifying, err := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		ServiceID:           "",
		SubscribeMerchantID: "",
		TransactionID:       "T3",
		ServiceType:         "",
		OrderType:           "",
		SubOrderType:        "",
		Amount:              11111,
		SourceOfFund:        "",
	})

	init_case_verifying_not_success, err := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		ServiceID:           "",
		SubscribeMerchantID: "",
		TransactionID:       "T3",
		ServiceType:         "",
		OrderType:           "",
		SubOrderType:        "",
		Amount:              11111,
		SourceOfFund:        "",
	})

	if err != nil || init_case1 == nil || init_case2 == nil {
		t.Error("can not init order")
	}

	type args struct {
		orderId string
		status  string
	}

	tests := []struct {
		name         string
		id           string
		expect       order_system.OrderStatus
		expectString string
		wantError    bool
		status       order_system.OrderStatus
		mockFunc     func()
		args
	}{
		{
			name:         "case 1 happy case",
			id:           "",
			expect:       order_system.OrderStatus_ORDER_SUCCESS,
			expectString: "ORDER_SUCCESS",
			wantError:    false,
			status:       order_system.OrderStatus_ORDER_SUCCESS,
			mockFunc: func() {
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: "T1",
				}, nil).Times(1)

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{
					ErrorCode: "",
					Data:      []*entities2.ListLinked{},
				}, nil)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

			},
			args: args{
				orderId: init_case1.OrderID,
				status:  constants.STATUS_SUCCESS,
			},
		},
		{
			name:         "case 2 fail case ",
			id:           "",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			wantError:    true,
			status:       order_system.OrderStatus_ORDER_FAILED,
			mockFunc: func() {
				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: "T",
				}, nil).Times(1)

				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{}, nil)

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

			},
			args: args{
				orderId: init_case2.OrderID,
				status:  constants.STATUS_FAILED,
			},
		},
		{
			name:         "case 3 verifying => retrieve Bank => success",
			id:           "",
			expect:       order_system.OrderStatus_ORDER_SUCCESS,
			expectString: "ORDER_SUCCESS",
			wantError:    false,
			status:       order_system.OrderStatus_ORDER_SUCCESS,
			mockFunc: func() {
				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{
					ErrorCode: "",
					Data:      []*entities2.ListLinked{},
				}, nil)

				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: "T3",
				}, nil).Times(1)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.BankService.On("RetrieveOrderStatus", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(entities2.RetrieveCreditOrderResponse{
					ErrorCode: "200",
				}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

			},
			args: args{
				orderId: init_case_verifying.OrderID,
				status:  constants.STATUS_VERIFYING,
			},
		},
		{
			name:         "case 4 verifying => retrieve Bank => not success",
			id:           "",
			expect:       order_system.OrderStatus_ORDER_VERIFYING,
			expectString: "ORDER_VERIFYING",
			wantError:    false,
			status:       order_system.OrderStatus_ORDER_VERIFYING,
			mockFunc: func() {
				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{
					ErrorCode: "",
					Data:      []*entities2.ListLinked{},
				}, nil)

				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: "T3",
				}, nil).Times(1)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.BankService.On("RetrieveOrderStatus", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(entities2.RetrieveCreditOrderResponse{
					ErrorCode: "400",
				}, nil).Times(1)

				th.Transaction.On("UpdateTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T3",
				}, nil)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T3",
				}, nil).Times(1)

			},
			args: args{
				orderId: init_case_verifying_not_success.OrderID,
				status:  constants.STATUS_VERIFYING,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			res := &order_system.UpdateCreditPaymentResponse{}

			order, err := th.OrderApplication.UpdateCreditPaymentOrder(ctx, &order_system.UpdateCreditPaymentRequest{
				OrderId: tt.orderId,
				Status:  tt.args.status,
			}, res)

			if err != nil {
				t.Log(err)
			}
			if !assert.Equal(t, tt.expect, order.Status.StatusOrderProto()) || !assert.Equal(t, tt.expectString, order.Status.StatusString()) {
				assert.Error(t, fmt.Errorf("status err order "))
			}

		})
	}
}

func TestOrderApplication_UpdateBalanceVA(t *testing.T) {
	th := NewTestOrderApplication()
	ctx := context.TODO()
	defer th.DB.Drop(ctx)

	init_case1, err := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		ServiceID:           "",
		SubscribeMerchantID: "",
		TransactionID:       "",
		ServiceType:         "",
		OrderType:           "",
		SubOrderType:        "",
		Amount:              99999,
		SourceOfFund:        "",
		ExpiredAt:           time.Now().Add(10 * time.Hour),
	})

	init_case2, err := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		ServiceID:           "",
		SubscribeMerchantID: "",
		TransactionID:       "",
		ServiceType:         "",
		OrderType:           "",
		SubOrderType:        "",
		Amount:              99999,
		SourceOfFund:        "",
		ExpiredAt:           time.Now().Add(10 * time.Hour),
	})

	if err != nil || init_case1 == nil || init_case2 == nil {
		t.Error("can not init order")
	}

	type args struct {
		req order_system.VAChangeBalanceResponse
	}

	tests := []struct {
		name         string
		id           string
		expect       order_system.OrderStatus
		expectString string
		wantError    bool
		status       order_system.OrderStatus
		mockFunc     func()
		args
	}{
		{
			name:         "case 1 happy case",
			id:           "",
			expect:       order_system.OrderStatus_ORDER_SUCCESS,
			expectString: "ORDER_SUCCESS",
			wantError:    false,
			status:       order_system.OrderStatus_ORDER_SUCCESS,
			mockFunc: func() {
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)
				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.IVA.On("IncrementBalanceVA", mock.Anything, mock.Anything).Return(entities.VirtualAccounts{
					MerchantId:   "1",
					MerchantCode: "1",
				}, nil)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.MerchantFeeService.On("CheckMerchantQuotaAndFee", ctx, mock.Anything).Return(&service_merchant_fee.CheckMerchantQuotaAndFeeRes{
					FeeAmount: 10,
					FeeMethod: 0,
				}, nil).Once()

			},
			args: args{
				req: order_system.VAChangeBalanceResponse{
					AccountNumber:     "99800020000000000101",
					Currency:          "VND",
					GpayAccountNumber: "",
					Amount:            "10000",
					BankTraceId:       "namnam1",
					PaymentMethod:     "CREDIT",
					BankTransactionId: "namnam1",
					Description:       "sdasd",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			request, err := proto.Marshal(&tt.args.req)
			if err != nil {
				panic(err)
			}
			order, err := th.OrderApplication.UpdateBalanceVA(ctx, request)

			if err != nil {
				t.Log(err)
			}
			if !assert.Equal(t, tt.expect, order.Status.StatusOrderProto()) || !assert.Equal(t, tt.expectString, order.Status.StatusString()) {
				assert.Error(t, fmt.Errorf("status err order "))
			}

		})
	}
}

func TestOrderApplication_BuyCardWithToken(t *testing.T) {
	type args struct {
		amount            int64
		quantity          int64
		telco             string
		voucherCode       string
		userId            string
		sourceOfFund      string
		confirmOTPRequest order_system.ConfirmPaymentTokenRequest
	}

	th := NewTestOrderApplication()
	ctx := context_grpc.NewOrderSystemContextGRPC(context.TODO())
	defer th.DB.Drop(ctx)

	paymentTokenOrderInit1, _ := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		ServiceID:    "123",
		UserID:       "123",
		ServiceType:  "123",
		OrderType:    "123",
		SubOrderType: "23",
		Amount:       222220,
		SourceOfFund: "xczxc",
		Status:       2,
		VoucherCode:  "asd",
		DeviceID:     "asd",
		BankCode:     "STB",
		PhoneTopUp:   "",
		ExpiredAt:    helpers.GetCurrentTime().Add(10 * time.Hour),
	})

	paymentTokenOrderInit, err := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		ServiceID:    "123",
		UserID:       "123",
		ServiceType:  "123",
		OrderType:    "123",
		SubOrderType: "23",
		Amount:       222220,
		SourceOfFund: "xczxc",
		Status:       2,
		VoucherCode:  "asd",
		DeviceID:     "asd",
		BankCode:     "STB",
		PhoneTopUp:   "",
		ExpiredAt:    helpers.GetCurrentTime().Add(10 * time.Hour),
	})

	if err != nil {
		panic(err)
	}

	tests := []struct {
		name         string
		id           string
		expect       order_system.OrderStatus
		expectString string
		wantError    bool
		status       order_system.OrderStatus
		fd           func()
		args
	}{
		{
			args: args{
				amount:       10000,
				quantity:     1,
				telco:        "VNP",
				voucherCode:  "Voucher",
				userId:       "U",
				sourceOfFund: constants.SOURCE_OF_FUND_BANK_ATM,
				confirmOTPRequest: order_system.ConfirmPaymentTokenRequest{
					BankTraceId: "qweqwe",
					OrderId:     paymentTokenOrderInit1.OrderID,
					LinkId:      1,
					Otp:         "123123",
				},
			},
			name:         "happy_case",
			expect:       order_system.OrderStatus_ORDER_SUCCESS,
			expectString: "ORDER_SUCCESS",
			status:       order_system.OrderStatus_ORDER_SUCCESS,
			fd: func() {
				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: "T1",
				}, nil)

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
				}, nil)

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.Promotion.On("UseVoucher", ctx, mock.Anything).Return(&service_promotion.UseVoucherResponse{
					DiscountAmount: 1,
					Voucher: &service_promotion.VoucherDetail{
						Voucher: &service_promotion.VoucherDTO{
							Id: "V",
						},
					},
				}, nil).Times(1)

				th.MerchantFeeService.On("GetMerchantVendorDiscount", ctx, mock.Anything).Return(&service_merchant_fee.GetMerchantVendorDiscountRes{
					MerchantDiscountAmount: 100,
					VendorDiscountAmount:   200,
				}, nil).Once()

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T1",
				}, nil).Times(1)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T1",
				}, nil).Times(1)

				th.Transaction.On("UpdateTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T1",
				}, nil)

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T1",
				}, nil)

				th.BankService.On("VerifyOTP", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities2.VerifyOTPResponse{
					ErrorCode: "200",
				}, nil).Times(1)

				th.Card.On("BuyCard", ctx, mock.Anything).Return(&service_card.BuyCardRes{
					Cards: []*service_card.CardObjDTO{{
						Provider:   "VTM",
						CardNumber: "CardNumber",
						Serial:     "Serial",
						Price:      "",
					}},
				}, nil).Times(1)

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

			},
			wantError: false,
		},
		{
			args: args{
				amount:       10000,
				quantity:     1,
				telco:        "VNP",
				userId:       "U2",
				sourceOfFund: constants.SOURCE_OF_FUND_BANK_ATM,
				confirmOTPRequest: order_system.ConfirmPaymentTokenRequest{
					BankTraceId: "casc",
					OrderId:     paymentTokenOrderInit.OrderID,
					LinkId:      0,
					Otp:         "ascasc",
				},
			},
			name:         "PAYBYTOKEN FAIL CASE  (FAIL VERIFY OTP)",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			status:       order_system.OrderStatus_ORDER_FAILED,
			fd: func() {
				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: "T",
				}, nil)

				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id: "U2",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 100000,
							AmountFreeze:    0,
						}},
					},
				}, nil).Times(5)

				th.Bank.On("GetUserLinkedList", mock.Anything).Return([]*entities.LinkedBankLink{
					{
						ID: "L",
					},
				}, nil).Times(2)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("UpdateTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

				th.BankService.On("VerifyOTP", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities2.VerifyOTPResponse{
					ErrorCode: "400",
				}, errors.New("wrong OTP")).Times(1)

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{}, nil)

				th.Card.On("BuyCard", ctx, mock.Anything).Return(&service_card.BuyCardRes{
					Cards: []*service_card.CardObjDTO{{
						Provider:   "VTM",
						CardNumber: "CardNumber",
						Serial:     "Serial",
						Price:      "",
					}},
				}, nil).Times(1)

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fd()
			res := &order_system.OrderBuyCardResponse{
				OrderEntity: nil,
				Cards:       nil,
			}

			order, err := th.OrderApplication.BuyCardWithToken(ctx, &order_system.OrderBuyCardRequest{
				Telco: tt.args.telco,
				OrderRequest: &order_system.OrderRequest{
					Amount:       tt.args.amount,
					Quantity:     tt.args.quantity,
					VoucherCode:  tt.args.voucherCode,
					UserID:       tt.args.userId,
					ServiceID:    constants.SUB_TRANSTYPE_WALLET_BUY_CARD,
					SubTransType: constants.SUB_TRANSTYPE_WALLET_BUY_CARD,
					SourceOfFund: tt.args.sourceOfFund,
				},
				ConfirmPaymentTokenRequest: &tt.args.confirmOTPRequest,
			}, res)

			if tt.wantError == false {
				if err != nil {
					assert.Error(t, err)
				}
				if !assert.Equal(t, tt.expect, order.Status.StatusOrderProto()) || !assert.Equal(t, tt.expectString, order.Status.StatusString()) {
					assert.Error(t, fmt.Errorf("status err order "))
				}

			} else {
				assert.NotEqual(t, err, nil)
			}
		})

	}

}

func TestOrderApplication_CashInNAPAS(t *testing.T) {
	th := NewTestOrderApplication()
	ctx := context.TODO()
	defer th.DB.Drop(ctx)

	init_case1, err := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		ServiceID:           "",
		SubscribeMerchantID: "",
		TransactionID:       "T",
		ServiceType:         "",
		OrderType:           "",
		SubOrderType:        "",
		Amount:              99999,
		SourceOfFund:        "",
		ExpiredAt:           helpers.GetCurrentTime().Add(10 * time.Hour),
	})

	init_case2, err := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		ServiceID:           "",
		SubscribeMerchantID: "",
		TransactionID:       "T1",
		ServiceType:         "",
		OrderType:           "",
		SubOrderType:        "",
		Amount:              99999,
		SourceOfFund:        "",
		ExpiredAt:           helpers.GetCurrentTime().Add(10 * time.Hour),
	})

	//init_case_verifying, err := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
	//	ServiceID:           "",
	//	SubscribeMerchantID: "",
	//	TransactionID:       "T3",
	//	ServiceType:         "",
	//	OrderType:           "",
	//	SubOrderType:        "",
	//	Amount:              11111,
	//	SourceOfFund:        "",
	//})
	//
	//init_case_verifying_not_success, err := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
	//	ServiceID:           "",
	//	SubscribeMerchantID: "",
	//	TransactionID:       "T3",
	//	ServiceType:         "",
	//	OrderType:           "",
	//	SubOrderType:        "",
	//	Amount:              11111,
	//	SourceOfFund:        "",
	//})

	if err != nil || init_case1 == nil || init_case2 == nil {
		t.Error("can not init order")
	}

	type args struct {
		orderId   string
		status    string
		errorCode string
		amount    int64
	}

	tests := []struct {
		name         string
		id           string
		expect       order_system.OrderStatus
		expectString string
		wantError    bool
		status       order_system.OrderStatus
		mockFunc     func()
		args
	}{
		{
			name:         "case 1 happy case",
			id:           "",
			expect:       order_system.OrderStatus_ORDER_SUCCESS,
			expectString: "ORDER_SUCCESS",
			wantError:    false,
			status:       order_system.OrderStatus_ORDER_SUCCESS,
			mockFunc: func() {
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: "T1",
				}, nil).Times(1)

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{
					ErrorCode: "",
					Data:      []*entities2.ListLinked{},
				}, nil)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

			},
			args: args{
				orderId:   init_case1.OrderID,
				status:    constants.STATUS_SUCCESS,
				errorCode: "200",
				amount:    99999,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			res := &order_system.CashInNapasResponse{}

			order, err := th.OrderApplication.CashInNAPAS(ctx, &order_system.CashInNapasRequest{
				OrderRequest: &order_system.OrderRequest{
					Amount:            tt.args.amount,
					BankTransactionId: tt.args.orderId,
				},
				ErrorCode: tt.args.errorCode,
			}, res)

			if err != nil {
				t.Log(err)
			}
			if !assert.Equal(t, tt.expect, order.Status.StatusOrderProto()) || !assert.Equal(t, tt.expectString, order.Status.StatusString()) {
				assert.Error(t, fmt.Errorf("status err order "))
			}

		})
	}
}

func TestOrderApplication_RefundTrans(t *testing.T) {
	th := NewTestOrderApplication()
	ctx := context.TODO()
	defer th.DB.Drop(ctx)

	initOrder_happy_case_merchant, err := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		ServiceID:           "",
		SubscribeMerchantID: "",
		TransactionID:       "T",
		ServiceType:         "",
		OrderType:           "",
		SubOrderType:        "",
		Amount:              99999,
		SourceOfFund:        "",
		ExpiredAt:           helpers.GetCurrentTime().Add(10 * time.Hour),
	})

	//initOrder_case_wallet, err := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
	//	ServiceID:           "",
	//	SubscribeMerchantID: "",
	//	TransactionID:       "T",
	//	ServiceType:         "",
	//	OrderType:           "",
	//	SubOrderType:        "",
	//	Amount:              99999,
	//	SourceOfFund:        "",
	//})

	if err != nil {
		panic(err)
	}

	type args struct {
		request *order_system.RefundTransactionReq
	}
	tests := []struct {
		name         string
		id           string
		expect       order_system.OrderStatus
		expectString string
		wantError    bool
		status       order_system.OrderStatus
		mockFunc     func()
		args
	}{
		{
			name:         "case 1 happy case",
			id:           "",
			expect:       order_system.OrderStatus_ORDER_SUCCESS,
			expectString: "ORDER_SUCCESS",
			wantError:    false,
			status:       order_system.OrderStatus_ORDER_SUCCESS,
			mockFunc: func() {
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id:     "T1",
					Status: constants.TRANSACTION_STATUS_FINISH,
				}, nil).Times(1)

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{
					ErrorCode: "",
					Data:      []*entities2.ListLinked{},
				}, nil)

				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: "T",
				}, nil).Times(1)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)
				th.Transaction.On("RefundTransaction", ctx, mock.Anything).Return(&service_transaction.RefundTransactionResponse{
					RefundTransactionId:    "refund-id",
					RefundFeeTransactionId: "refund-id",
					RefundId:               "refund-id",
				}, nil).Times(1)

				th.Transaction.On("ConfirmRefundTransaction", ctx, &service_transaction.ConfirmRefundTransactionRequest{
					RefundId:       "refund-id",
					ConfirmByAdmin: "Hệ thống",
				}).Return(&service_transaction.RefundTransactionResponse{
					RefundId: "refund-id-confirm",
				}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

			},
			args: args{
				request: &order_system.RefundTransactionReq{
					SourceTransactionID: "transaction1",
					Amount:              99999,
					Reason:              "test",
					Note:                "test",
					SourceOrderId:       initOrder_happy_case_merchant.OrderID,
					MerchantId:          "",
					RefundType:          order_system.RefundTransactionReq_INDIRECT,
				},
			},
		},
		{
			name:         "case 2 FAIL case",
			id:           "",
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			wantError:    false,
			status:       order_system.OrderStatus_ORDER_FAILED,
			mockFunc: func() {
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: "T1",
				}, nil).Times(1)

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{
					ErrorCode: "",
					Data:      []*entities2.ListLinked{},
				}, nil)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)
				th.Transaction.On("RefundTransaction", ctx, mock.Anything).Return(&service_transaction.RefundTransactionResponse{
					RefundTransactionId:    "transaction-refund-1",
					RefundFeeTransactionId: "transaction-fee-1",
					RefundId:               "refund-id",
				}, errors.New("số tiền hoàn lớn hơn GD gốc")).Times(1)
				th.Transaction.On("CancelRefundTransaction", ctx, mock.Anything).Return(&service_transaction.RefundTransactionResponse{}, nil).Once()

			},
			args: args{
				request: &order_system.RefundTransactionReq{
					SourceTransactionID: "transaction1",
					Amount:              99999,
					Reason:              "test",
					Note:                "test",
				},
			},
		},
		{
			name:         "case 3 FAIL case (MERCHANT ko du tien STATUS_SETTLEMENT == SUCCESS)", // check amount_cash
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			wantError:    false,
			status:       order_system.OrderStatus_ORDER_FAILED,
			mockFunc: func() {
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.Transaction.On("RefundTransaction", ctx, mock.Anything).Return(&service_transaction.RefundTransactionResponse{
					RefundTransactionId:    "transaction-refund-1",
					RefundFeeTransactionId: "transaction-fee-1",
					RefundId:               "refund-id",
				}, nil).Times(1)

				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id:               "T1",
					Status:           constants.TRANSACTION_STATUS_FINISH,
					StatusSettlement: constants.STATUS_SUCCESS,
					MerchantID:       "merchant-1",
				}, nil).Times(1)

				th.User.On("GetMerchantAccount", ctx, mock.Anything).Return(&service_user.MerchantAccount{
					MerchantId:    "",
					AmountRevenue: 99,
					AmountCash:    0,
				}, nil).Once()

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{
					ErrorCode: "",
					Data:      []*entities2.ListLinked{},
				}, nil)

				th.Transaction.On("CancelRefundTransaction", ctx, mock.Anything).Return(&service_transaction.RefundTransactionResponse{}, nil).Once()

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

			},
			args: args{
				request: &order_system.RefundTransactionReq{
					SourceTransactionID: "T1",
					Amount:              99999,
					Reason:              "test",
					Note:                "test",
				},
			},
		},
		{
			name:         "case 4 FAIL case (MERCHANT ko du tien case STATUS_SETTLEMENT != SUCCESS)", // check amount_revenue (trade)
			expect:       order_system.OrderStatus_ORDER_FAILED,
			expectString: "ORDER_FAILED",
			wantError:    false,
			status:       order_system.OrderStatus_ORDER_FAILED,
			mockFunc: func() {
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.User.On("GetMerchantAccount", ctx, mock.Anything).Return(&service_user.MerchantAccount{
					MerchantId:    "",
					AmountRevenue: 99,
					AmountCash:    0,
				}, nil).Once()

				th.Transaction.On("RefundTransaction", ctx, mock.Anything).Return(&service_transaction.RefundTransactionResponse{
					RefundTransactionId:    "transaction-refund-1",
					RefundFeeTransactionId: "transaction-fee-1",
					RefundId:               "refund-id",
				}, nil).Times(1)

				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id:         "T1",
					Status:     constants.TRANSACTION_STATUS_FINISH,
					MerchantID: "merchant-1",
				}, nil).Times(1)

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{
					ErrorCode: "",
					Data:      []*entities2.ListLinked{},
				}, nil)

				th.Transaction.On("CancelRefundTransaction", ctx, mock.Anything).Return(&service_transaction.RefundTransactionResponse{}, nil).Once()

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

			},
			args: args{
				request: &order_system.RefundTransactionReq{
					SourceTransactionID: "T1",
					Amount:              99999,
					Reason:              "test",
					Note:                "test",
				},
			},
		},
		{
			name:         "case 5 refund config not AUTO",
			expect:       order_system.OrderStatus_ORDER_PROCESSING,
			expectString: "ORDER_PROCESSING",
			wantError:    false,
			status:       order_system.OrderStatus_ORDER_PROCESSING,
			mockFunc: func() {
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.User.On("GetMerchantAccount", ctx, mock.Anything).Return(&service_user.MerchantAccount{
					MerchantId:    "namle123",
					AmountRevenue: 11111111,
					AmountCash:    1111111,
				}, nil).Once()

				th.Transaction.On("RefundTransaction", ctx, mock.Anything).Return(&service_transaction.RefundTransactionResponse{
					RefundTransactionId:    "transaction-refund-1",
					RefundFeeTransactionId: "transaction-fee-1",
					RefundId:               "refund-id",
				}, nil).Times(1)

				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id:               "T1",
					Status:           constants.TRANSACTION_STATUS_FINISH,
					StatusSettlement: constants.STATUS_SUCCESS,
					TransactionType:  "PAY",
					SourceOfFund:     constants.SOURCE_OF_FUND_BANK_ATM,
				}, nil).Times(1)

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{
					ErrorCode: "",
					Data:      []*entities2.ListLinked{},
				}, nil)

				th.Transaction.On("CancelRefundTransaction", ctx, mock.Anything).Return(&service_transaction.RefundTransactionResponse{}, nil).Once()

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.IWalletConfig.On("GetRefundConfig", ctx, request_params.GetRefundConfigReq{
					MerchantId:   "namle123",
					TransType:    "PAY",
					SourceOfFund: constants.SOURCE_OF_FUND_BANK_ATM,
				}).Return(value_objects.GetRefundConfigRes{
					TransType: "PAY",
					Settings: []value_objects.RefundSetting{{
						SourceOfFunds:   []string{constants.SOURCE_OF_FUND_BANK_ATM},
						PaymentStatus:   "UNPAID",
						RefundType:      "INDIRECT",
						ConfirmType:     "",
						RefundCondition: "LE",
						RefundValue:     111111,
						Status:          "ACTIVE",
					}, {
						SourceOfFunds:   []string{constants.SOURCE_OF_FUND_BANK_ATM},
						PaymentStatus:   "PAID",
						RefundType:      "INDIRECT",
						ConfirmType:     "",
						RefundCondition: "LTE",
						RefundValue:     0,
						Status:          "",
					}},
				}, nil).Once()

			},
			args: args{
				request: &order_system.RefundTransactionReq{
					SourceTransactionID: "T1",
					Amount:              111,
					Reason:              "test",
					Note:                "test",
					MerchantId:          "namle123",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			res, err := th.OrderApplication.InitRefundTrans(ctx, &order_system.RefundTransactionReq{
				SourceTransactionID: "refund-tran-1",
				Amount:              tt.args.request.Amount,
				Reason:              "GD bi loi",
				Note:                "hoan tien nha",
				SourceOrderId:       tt.args.request.SourceOrderId,
				MerchantId:          tt.args.request.MerchantId,
			})

			if err != nil {
				fmt.Println("err", err)
			}

			if tt.wantError {
				assert.Error(t, err)
			} else {
				if !assert.Equal(t, tt.expect, res.Status.StatusOrderProto()) || !assert.Equal(t, tt.expectString, res.Status.StatusString()) {
					assert.Error(t, fmt.Errorf("different order status"))
				}
			}

		})
	}
}

func TestOrderApplication_PayInternationalCard(t *testing.T) {
	//th := NewTestOrderApplication()
	//ctx := context.TODO()
	//defer th.DB.Drop(ctx)
	//
	//initOrder_happy_case, err := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
	//	ServiceID:           "",
	//	SubscribeMerchantID: "",
	//	TransactionID:       "T",
	//	ServiceType:         "",
	//	OrderType:           "",
	//	SubOrderType:        "",
	//	Amount:              99999,
	//	SourceOfFund:        "",
	//	ExpiredAt:           helpers.GetCurrentTime().Add(10 * time.Hour),
	//})
	//
	//
	//
	//type args struct {
	//	request *order_system.PayInternationalCardRequest
	//}
	//tests := []struct {
	//	name         string
	//	id           string
	//	expect       order_system.OrderStatus
	//	expectString string
	//	wantError    bool
	//	status       order_system.OrderStatus
	//	mockFunc     func()
	//	args
	//}{
	//	{
	//		name:         "case 1 happy case",
	//		id:           "",
	//		expect:       order_system.OrderStatus_ORDER_SUCCESS,
	//		expectString: "ORDER_SUCCESS",
	//		wantError:    false,
	//		status:       order_system.OrderStatus_ORDER_SUCCESS,
	//		mockFunc: func() {
	//			th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
	//				UserDetail: &service_user.UserDetailDTO{
	//					User: &service_user.UserDTO{
	//						Id:                "user-qwe1",
	//						HasEverLinkedBank: true,
	//						Kyc:               "ACTIVE",
	//					},
	//					Balances: []*service_user.BalanceDTO{{
	//						AmountAvailable: 1000010,
	//						AmountFreeze:    0,
	//					}},
	//				},
	//			}, nil)
	//
	//			th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
	//				Id:     "T1",
	//				Status: constants.TRANSACTION_STATUS_FINISH,
	//			}, nil).Times(1)
	//
	//			th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything ).Return(nil)
	//
	//			th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{
	//				ErrorCode: "",
	//				Data:      []*entities2.ListLinked{},
	//			}, nil)
	//
	//			th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
	//				Id: "T",
	//			}, nil).Times(1)
	//
	//			th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)
	//			th.Transaction.On("RefundTransaction", ctx, mock.Anything).Return(&service_transaction.RefundTransactionResponse{
	//				RefundTransactionId:    "refund-id",
	//				RefundFeeTransactionId: "refund-id",
	//				RefundId:               "refund-id",
	//			}, nil).Times(1)
	//
	//			th.User.On("GetMerchantAccount", ctx, mock.Anything).Return(&service_user.MerchantAccount{
	//				MerchantId:    "",
	//				AmountRevenue: 10000000,
	//				AmountCash:    10000000,
	//			}, nil).Once()
	//
	//			th.Transaction.On("ConfirmRefundTransaction", ctx, &service_transaction.ConfirmRefundTransactionRequest{
	//				RefundId:       "refund-id",
	//				ConfirmByAdmin: "Hệ thống",
	//			}).Return(&service_transaction.RefundTransactionResponse{
	//				RefundId: "refund-id-confirm",
	//			}, nil).Times(1)
	//
	//			th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
	//				TransactionId: "T",
	//			}, nil).Times(1)
	//
	//			th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
	//				TransactionId: "T",
	//			}, nil).Times(1)
	//
	//		},
	//		args: args{
	//			request: &order_system.PayInternationalCardRequest{
	//				GpayBankCode:         "",
	//				GpayUserId:           "",
	//				Amount:               0,
	//				LinkId:               0,
	//				TransType:            "",
	//				SubTransType:         "",
	//				VoucherCode:          "",
	//				ServiceCode:          "",
	//				Quantity:             0,
	//				Telco:                "",
	//				ServiceCodeBill:      "",
	//				CustomerReference:    "",
	//				AreaCode:             "",
	//				OrderId:              "",
	//			},
	//		},
	//	},
	//}
	//
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//		tt.mockFunc()
	//		res, err := th.OrderApplication.PayInternationalCard(ctx, &order_system.PayInternationalCardRequest{
	//			GpayBankCode:         "",
	//			GpayUserId:           "",
	//			Amount:               tt.args.request.Amount,
	//			LinkId:               0,
	//			TransType:            "",
	//			SubTransType:         "",
	//			VoucherCode:          "",
	//			ServiceCode:          "",
	//			Quantity:             0,
	//			Telco:                "",
	//			ServiceCodeBill:      "",
	//			CustomerReference:    "",
	//			AreaCode:             "",
	//			OrderId:              "",
	//		})
	//
	//		if err != nil {
	//			fmt.Println("err", err)
	//		}
	//
	//		if tt.wantError {
	//			assert.Error(t, err)
	//		} else {
	//			if !assert.Equal(t, tt.expect, res.Status.StatusOrderProto()) || !assert.Equal(t, tt.expectString, res.Status.StatusString()) {
	//				assert.Error(t, fmt.Errorf("different order status"))
	//			}
	//		}
	//
	//	})
	//}

}

func TestOrderApplication_UpdateBankStatus(t *testing.T) {
	th := NewTestOrderApplication()
	ctx := context.TODO()
	defer th.DB.Drop(ctx)

	orderCashInHappy, _ := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		UserID:       "123",
		OrderType:    "CASHIN",
		SubOrderType: "23",
		Amount:       222220,
		SourceOfFund: "xczxc",
		Status:       1,
		ExpiredAt:    helpers.GetCurrentTime().Add(10 * time.Hour),
	})

	qrWEBINAPP, _ := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		UserID:       "123",
		OrderType:    "CASHIN",
		SubOrderType: constants.SUB_TRANSTYPE_WALLET_WEB_IN_APP,
		Amount:       222220,
		SourceOfFund: "xczxc",
		Status:       1,
	})

	orderCashInFail, _ := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		UserID:       "123",
		OrderType:    "CASHIN",
		SubOrderType: "23",
		Amount:       222220,
		SourceOfFund: "xczxc",
		Status:       1,
		ExpiredAt:    helpers.GetCurrentTime().Add(10 * time.Hour),
	})

	orderBUYCARDHappy, _ := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		UserID:         "123",
		OrderType:      constants.TRANSTYPE_WALLET_PAY,
		SubOrderType:   constants.SUB_TRANSTYPE_WALLET_BUY_CARD,
		Amount:         20000,
		SourceOfFund:   constants.SOURCE_OF_FUND_INTERNATIONAL_CARD,
		Quantity:       2,
		OrderCardTelco: "VNP",
		ExpiredAt:      helpers.GetCurrentTime().Add(10 * time.Hour),
	})

	orderTOPUPHappycase, _ := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		UserID:       "123",
		OrderType:    constants.TRANSTYPE_WALLET_PAY,
		SubOrderType: constants.SUB_TRANSTYPE_WALLET_TOPUP_CARD,
		Amount:       20000,
		SourceOfFund: constants.SOURCE_OF_FUND_INTERNATIONAL_CARD,
		ExpiredAt:    helpers.GetCurrentTime().Add(10 * time.Hour),
	})

	orderPAIDBillHappycase, _ := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		UserID:               "123",
		OrderType:            constants.TRANSTYPE_WALLET_PAY,
		SubOrderType:         constants.SUB_TRANSTYPE_WALLET_PAY_BILL_ELECTRIC,
		Amount:               20000,
		OrderBillAreaCode:    "area-bill",
		OrderBillCustomerRef: "customer-bill",
		OrderBillServiceCode: "service-bill",
		SourceOfFund:         constants.SOURCE_OF_FUND_INTERNATIONAL_CARD,
		ExpiredAt:            helpers.GetCurrentTime().Add(10 * time.Hour),
	})

	orderPAIDBillFailcase, _ := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		UserID:               "123",
		OrderType:            constants.TRANSTYPE_WALLET_PAY,
		SubOrderType:         constants.SUB_TRANSTYPE_WALLET_PAY_BILL_ELECTRIC,
		Amount:               20000,
		OrderBillAreaCode:    "area-bill",
		OrderBillCustomerRef: "customer-bill",
		OrderBillServiceCode: "service-bill",
		SourceOfFund:         constants.SOURCE_OF_FUND_INTERNATIONAL_CARD,
		ExpiredAt:            helpers.GetCurrentTime().Add(10 * time.Hour),
	})

	orderPAIDBillVerifyingCase, _ := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		UserID:               "123",
		OrderType:            constants.TRANSTYPE_WALLET_PAY,
		SubOrderType:         constants.SUB_TRANSTYPE_WALLET_PAY_BILL_ELECTRIC,
		Amount:               20000,
		OrderBillAreaCode:    "area-bill",
		OrderBillCustomerRef: "customer-bill",
		OrderBillServiceCode: "service-bill",
		SourceOfFund:         constants.SOURCE_OF_FUND_INTERNATIONAL_CARD,
		ExpiredAt:            helpers.GetCurrentTime().Add(10 * time.Hour),
	})

	failCaseBankStatus, _ := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		UserID:               "123",
		OrderType:            constants.TRANSTYPE_WALLET_PAY,
		SubOrderType:         constants.SUB_TRANSTYPE_WALLET_PAY_BILL_ELECTRIC,
		Amount:               20000,
		OrderBillAreaCode:    "area-bill",
		OrderBillCustomerRef: "customer-bill",
		OrderBillServiceCode: "service-bill",
		SourceOfFund:         constants.SOURCE_OF_FUND_INTERNATIONAL_CARD,
		ExpiredAt:            helpers.GetCurrentTime().Add(10 * time.Hour),
	})

	type args struct {
		ctx context.Context
		req order_system.BankOrderStatusResponse
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		expect       order_system.OrderStatus
		expectString string
		wantError    bool
		status       order_system.OrderStatus
		mockFunc     func()
	}{
		{
			name: "CASHIN (happy case)",
			args: args{
				ctx: ctx,
				req: order_system.BankOrderStatusResponse{
					OrderId:           orderCashInHappy.OrderID,
					Amount:            "10000",
					RefId:             "ref",
					BankTransactionId: "banktrace",
					ErrorCode:         "200",
					Description:       "nam",
				},
			},
			wantErr: false,
			mockFunc: func() {
				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: "T",
				}, nil).Once()

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1).Once()

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.MerchantFeeService.On("CheckMerchantQuotaAndFee", ctx, mock.Anything).Return(&service_merchant_fee.CheckMerchantQuotaAndFeeRes{
					FeeAmount: 10,
					FeeMethod: 0,
				}, nil).Once()

			},
			expect: order_system.OrderStatus_ORDER_SUCCESS,
		},
		{
			name: "CASHIN (fail case)",
			args: args{
				ctx: ctx,
				req: order_system.BankOrderStatusResponse{
					OrderId:           orderCashInFail.OrderID,
					Amount:            "10000",
					RefId:             "ref",
					BankTransactionId: "banktrace",
					ErrorCode:         "400",
					Description:       "nam",
				},
			},
			wantErr: false,
			mockFunc: func() {
				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: "T",
				}, nil).Once()

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.Transaction.On("UpdateTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1).Once()

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1).Once()

			},
			expect: order_system.OrderStatus_ORDER_FAILED,
		},
		{
			name: "BUYCARD (happy case)",
			args: args{
				ctx: ctx,
				req: order_system.BankOrderStatusResponse{
					OrderId:           orderBUYCARDHappy.OrderID,
					RefId:             "ref",
					BankTransactionId: "banktrace",
					ErrorCode:         "200",
					Description:       "nam",
				},
			},
			wantErr: false,
			mockFunc: func() {
				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: "T",
				}, nil).Once()

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1).Once()

				th.Card.On("BuyCard", ctx, mock.Anything).Return(&service_card.BuyCardRes{
					Status:             constants.STATUS_SUCCESS,
					ProviderMerchant:   "MERCHANT1",
					ProviderMerchantId: "MERCHANT1",
				}, nil)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1).Once()

				th.MerchantFeeService.On("GetMerchantVendorDiscount", ctx, mock.Anything).Return(&service_merchant_fee.GetMerchantVendorDiscountRes{
					MerchantDiscountAmount: 100,
					VendorDiscountAmount:   200,
				}, nil).Once()

			},
			expect: order_system.OrderStatus_ORDER_SUCCESS,
		},
		{
			name: "TOPUP (happy case)",
			args: args{
				ctx: ctx,
				req: order_system.BankOrderStatusResponse{
					OrderId:           orderTOPUPHappycase.OrderID,
					RefId:             "ref",
					BankTransactionId: "banktrace",
					ErrorCode:         "200",
					Description:       "nam",
				},
			},
			wantErr: false,
			mockFunc: func() {
				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: "T",
				}, nil).Once()

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1).Once()

				th.Card.On("Topup", ctx, mock.Anything).Return(&service_card.TopupRes{
					Status:             constants.STATUS_SUCCESS,
					ProviderMerchant:   "MERCHANT1",
					ProviderMerchantId: "MERCHANT1",
				}, nil)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1).Once()

			},
			expect: order_system.OrderStatus_ORDER_SUCCESS,
		},
		{
			name: "PAID BILL (happy case)",
			args: args{
				ctx: ctx,
				req: order_system.BankOrderStatusResponse{
					OrderId:           orderPAIDBillHappycase.OrderID,
					RefId:             "ref",
					BankTransactionId: "banktrace",
					ErrorCode:         "200",
					Description:       "nam",
				},
			},
			wantErr: false,
			mockFunc: func() {
				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: "T",
				}, nil).Once()

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Card.On("PaidBill", ctx, mock.Anything).Return(&service_card.PaidBillRes{
					Status:             constants.STATUS_SUCCESS,
					ProviderMerchant:   "MERCHANT1",
					ProviderMerchantId: "MERCHANT1",
				}, nil).Once()
				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1).Once()

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1).Once()

			},
			expect: order_system.OrderStatus_ORDER_SUCCESS,
		},
		{
			name: "PAID BILL (fail case)",
			args: args{
				ctx: ctx,
				req: order_system.BankOrderStatusResponse{
					OrderId:           orderPAIDBillFailcase.OrderID,
					RefId:             "ref",
					BankTransactionId: "banktrace",
					ErrorCode:         "200",
					Description:       "nam",
				},
			},
			wantErr: false,
			mockFunc: func() {
				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: "T",
				}, nil).Once()

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)
				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{}, nil)

				th.Card.On("PaidBill", ctx, mock.Anything).Return(&service_card.PaidBillRes{
					Status:             constants.STATUS_FAILED,
					ProviderMerchant:   "MERCHANT1",
					ProviderMerchantId: "MERCHANT1",
				}, errors.New("fail paid bill")).Once()
				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1).Once()

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1).Once()

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

			},
			expect: order_system.OrderStatus_ORDER_FAILED,
		},
		{
			name: "PAID BILL (verifying case)",
			args: args{
				ctx: ctx,
				req: order_system.BankOrderStatusResponse{
					OrderId:           orderPAIDBillVerifyingCase.OrderID,
					RefId:             "ref",
					BankTransactionId: "banktrace",
					ErrorCode:         "200",
					Description:       "nam",
				},
			},
			wantErr: false,
			mockFunc: func() {
				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: "T",
				}, nil).Once()

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)
				th.Promotion.On("ReverseWallet", ctx, mock.Anything).Return(&service_promotion.ReverseWalletRequest{}, nil)

				th.Card.On("PaidBill", ctx, mock.Anything).Return(&service_card.PaidBillRes{
					Status:             constants.TRANSACTION_STATUS_PENDING,
					ProviderMerchant:   "MERCHANT1",
					ProviderMerchantId: "MERCHANT1",
				}, nil).Once()
				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1).Once()

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1).Once()

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

			},
			expect: order_system.OrderStatus_ORDER_VERIFYING,
		},
		{
			name: "fail case (BANK STATUS != 200)",
			args: args{
				ctx: ctx,
				req: order_system.BankOrderStatusResponse{
					OrderId:           failCaseBankStatus.OrderID,
					Amount:            "10000",
					RefId:             "ref",
					BankTransactionId: "banktrace",
					ErrorCode:         "102",
					Description:       "nam",
				},
			},
			wantErr: false,
			mockFunc: func() {
				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: "T",
				}, nil).Once()

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.Transaction.On("UpdateTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1).Once()

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

			},
			expect: order_system.OrderStatus_ORDER_FAILED,
		},
		{
			name: "QR WEBINAPP (happy case)",
			args: args{
				ctx: ctx,
				req: order_system.BankOrderStatusResponse{
					OrderId:           qrWEBINAPP.OrderID,
					Amount:            "10000",
					RefId:             "ref",
					BankTransactionId: "banktrace",
					ErrorCode:         "200",
					Description:       "nam",
				},
			},
			wantErr: false,
			mockFunc: func() {
				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id:                 "T",
					SubTransactionType: constants.SUB_TRANSTYPE_WALLET_WEB_IN_APP,
				}, nil).Once()

				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId:      "T",
					SubTransactionType: constants.SUB_TRANSTYPE_WALLET_WEB_IN_APP,
				}, nil).Times(1).Once()

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId:      "Tqqqqqqqqqqqqqqqqqqq",
					SubTransactionType: constants.SUB_TRANSTYPE_WALLET_WEB_IN_APP,
				}, nil).Times(1)

				th.Transaction.On("CancelTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.MerchantFeeService.On("CheckMerchantQuotaAndFee", ctx, mock.Anything).Return(&service_merchant_fee.CheckMerchantQuotaAndFeeRes{
					FeeAmount: 10,
					FeeMethod: 0,
				}, nil).Once()

			},
			expect: order_system.OrderStatus_ORDER_SUCCESS,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			req := order_system.BankOrderStatusResponse{
				OrderId:           tt.args.req.OrderId,
				Amount:            tt.args.req.Amount,
				RefId:             tt.args.req.RefId,
				BankTransactionId: tt.args.req.BankTransactionId,
				ErrorCode:         tt.args.req.ErrorCode,
				Description:       tt.args.req.Description,
			}

			b, err := proto.Marshal(&req)
			if err != nil {
				panic(err)
			}

			order, err := th.OrderApplication.UpdateBankStatus(ctx, b)
			if err != nil {
				t.Log(err)
			}
			if !assert.Equal(t, tt.expect, order.Status.StatusOrderProto()) {
				assert.Error(t, fmt.Errorf("status err order "))
			}
		})
	}
}

func TestOrderApplication_StaticQRWithToken(t *testing.T) {
	th := NewTestOrderApplication()
	ctx := context.TODO()
	defer th.DB.Drop(ctx)

	init_case1, err := th.OrderApplication.InitOrder(ctx, &entities.OrderEntity{
		ServiceID:           "",
		UserID:              "namnam",
		SubscribeMerchantID: "",
		TransactionID:       "",
		ServiceType:         "",
		OrderType:           "",
		SubOrderType:        "",
		Amount:              99999,
		SourceOfFund:        "",
	})

	if err != nil || init_case1 == nil {
		panic(err)
	}

	tests := []struct {
		name         string
		id           string
		expect       order_system.OrderStatus
		expectString string
		wantError    bool
		status       order_system.OrderStatus
		mockFunc     func()
		args         order_system.PaymentMerchantRequest
	}{
		{
			name:         "case 1  By Pass OTP",
			id:           "",
			expect:       order_system.OrderStatus_ORDER_SUCCESS,
			expectString: "ORDER_SUCCESS",
			wantError:    false,
			status:       order_system.OrderStatus_ORDER_SUCCESS,
			mockFunc: func() {
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)
				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.BankService.On("LinkInfo", mock.AnythingOfType("int64")).Return(entities2.LinkInfoResponse{
					ErrorCode: "000",
					Message:   "",
				}, nil)

				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{
					ErrorCode: "",
					Data:      []*entities2.ListLinked{},
				}, nil)

				th.BankService.On("CashIn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities2.CashInResponse{
					ErrorCode: "200",
				}, nil).Once()

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

			},
			args: order_system.PaymentMerchantRequest{
				OrderRequest: &order_system.OrderRequest{
					Amount: 1,
				},
				ConfirmPaymentTokenRequest: &order_system.ConfirmPaymentTokenRequest{
					OrderId: init_case1.OrderID,
					LinkId:  0,
					Otp:     "",
				},
			},
		},
		{
			name:         "case 2 Not BY Pass OTP",
			id:           "",
			expect:       order_system.OrderStatus_ORDER_SUCCESS,
			expectString: "ORDER_SUCCESS",
			wantError:    false,
			status:       order_system.OrderStatus_ORDER_SUCCESS,
			mockFunc: func() {
				th.User.On("FindUserDetailById", ctx, mock.Anything).Return(&service_user.FindUserDetailByIdResponse{
					UserDetail: &service_user.UserDetailDTO{
						User: &service_user.UserDTO{
							Id:                "user-qwe1",
							HasEverLinkedBank: true,
							Kyc:               "ACTIVE",
						},
						Balances: []*service_user.BalanceDTO{{
							AmountAvailable: 1000010,
							AmountFreeze:    0,
						}},
					},
				}, nil)
				th.Mqtt.On("Publish", mock.Anything, mock.Anything, false, mock.Anything).Return(nil)

				th.BankService.On("LinkInfo", mock.AnythingOfType("int64")).Return(entities2.LinkInfoResponse{
					ErrorCode: "000",
					Message:   "",
				}, nil)

				th.BankService.On("LinkList", mock.Anything).Return(entities2.ListLinkResponse{
					ErrorCode: "",
					Data:      []*entities2.ListLinked{},
				}, nil)

				th.Transaction.On("FindTransactionByID", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					Id: "T",
				}, nil)

				th.BankService.On("VerifyOTP", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities2.VerifyOTPResponse{}, nil)

				th.BankService.On("CashIn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities2.CashInResponse{
					ErrorCode: "200",
				}, nil).Once()

				th.Transaction.On("CheckTransactionQuotaAndFee", ctx, mock.Anything).Return(&service_transaction.CheckTransactionQuotaAndFeeRes{}, nil).Times(1)

				th.Transaction.On("InitTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

				th.Transaction.On("ConfirmTransaction", ctx, mock.Anything).Return(&service_transaction.ETransactionDTO{
					TransactionId: "T",
				}, nil).Times(1)

			},
			args: order_system.PaymentMerchantRequest{
				OrderRequest: &order_system.OrderRequest{
					Amount: 1,
				},
				ConfirmPaymentTokenRequest: &order_system.ConfirmPaymentTokenRequest{
					OrderId: init_case1.OrderID,
					LinkId:  0,
					Otp:     "1234231",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			res := &order_system.PaymentMerchantResponse{}

			order, err := th.OrderApplication.StaticQRWithToken(ctx, &tt.args, res)

			if err != nil {
				t.Log(err)
			}
			if !assert.Equal(t, order.Status.StatusOrderProto(), tt.expect) || !assert.Equal(t, order.Status.StatusString(), tt.expectString) {
				assert.Error(t, fmt.Errorf("status err order "))
			}

		})
	}

}
