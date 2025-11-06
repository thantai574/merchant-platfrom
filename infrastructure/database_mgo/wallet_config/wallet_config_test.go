package wallet_config_test

import (
	"context"
	"orders-system/domain/request_params"
	"orders-system/infrastructure/database_mgo"
	"orders-system/infrastructure/database_mgo/wallet_config"
	"testing"
)

var uri string
var contextTest context.Context

func init() {
	uri = "mongodb://root:Vietnam2020@1.55.214.191:20003/"
	contextTest = context.TODO()
}
func TestRepoImpl_GetMerchantConfig(t *testing.T) {
	type args struct {
		ctx context.Context
		req request_params.GetMerchantConfigReq
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				ctx: contextTest,
				req: request_params.GetMerchantConfigReq{
					MerchantId:   "8013c15d-2307-4c04-8a15-7296f3ec7cc8",
					ServiceType:  "PAY_COLLECT",
					TransType:    "VA",
					SubTransType: "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := wallet_config.NewServiceWalletConfigRepository(database_mgo.NewMongoDBconnection(uri), "wallet", nil)
			got, err := r.GetMerchantConfig(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMerchantConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log(got)
		})
	}
}

func TestRepoImpl_GetRefundConfig(t *testing.T) {
	type args struct {
		ctx context.Context
		req request_params.GetRefundConfigReq
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				ctx: contextTest,
				req: request_params.GetRefundConfigReq{
					TransType:  "PAY",
					MerchantId: "23eff3ad-b788-425a-bcc1-a09a3a98a170",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := wallet_config.NewServiceWalletConfigRepository(database_mgo.NewMongoDBconnection(uri), "wallet", nil)
			got, err := r.GetRefundConfig(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMerchantConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log(got)
		})
	}
}
