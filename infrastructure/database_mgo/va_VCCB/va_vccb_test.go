package va_VCCB_test

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"orders-system/domain/entities"
	"orders-system/infrastructure/database_mgo"
	"orders-system/infrastructure/database_mgo/va_VCCB"
	"orders-system/utils/configs"
	"orders-system/utils/helpers"
	"testing"
)

func InitMongoDB() (*mongo.Client, error) {
	uri := "mongodb://root:Vietnam2020@1.55.214.191:20003/"
	return database_mgo.NewMongoDBconnection(uri), nil
}

const dbName = "va_vccb"

func TestRepoImpl_incrementID(t *testing.T) {
	db, err := InitMongoDB()
	if err != nil {
		t.Error(err)
		return
	}

	type args struct {
		ctx      context.Context
		prefix   string
		provider string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				ctx:      context.TODO(),
				prefix:   "XXX",
				provider: "MSB",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := va_VCCB.NewVAVCCBRepository(db, dbName, &configs.Config{})
			got, err := r.IncrementID(tt.args.ctx, tt.args.prefix, tt.args.provider)
			if (err != nil) != tt.wantErr {
				t.Errorf("incrementID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Log("got", got)

		})
	}
}

func TestRepoImpl_CreateVA(t *testing.T) {
	db, err := InitMongoDB()
	if err != nil {
		t.Error(err)
		return
	}

	type args struct {
		request entities.VirtualAccounts
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				request: entities.VirtualAccounts{
					IsAutoIncrement: true,
					AccountType:     "ONETIME",
					CreatedAt:       helpers.GetCurrentTime(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := va_VCCB.NewVAVCCBRepository(db, dbName, &configs.Config{})
			gotRes, err := r.CreateVA(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateVA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log("gotRes", gotRes)
		})
	}
}

func TestRepoImpl_DeleteVAAccount(t *testing.T) {
	db, err := InitMongoDB()
	if err != nil {
		t.Error(err)
		return
	}

	type args struct {
		accNumber string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				accNumber: "M010200000000044",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := va_VCCB.NewVAVCCBRepository(db, dbName, &configs.Config{})
			err := r.DeleteVAAccount(tt.args.accNumber)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateVA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log(err)
		})
	}
}
