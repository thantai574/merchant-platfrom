package lixi

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"orders-system/domain/entities"
	"orders-system/infrastructure/database_mgo"
	"testing"
)

func TestLixiCollection_FindStarted(t *testing.T) {
	type fields struct {
		collection       *mongo.Collection
		collectionAmount *mongo.Collection
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantRes  []*entities.LixiEntity
		wantErr  bool
		lenGotGt int
	}{
		{
			name: "find all started",
			fields: fields{
				collection:       database_mgo.NewMongoDBconnection("mongodb://root:Vietnam2020@1.55.214.191:20003/").Database("service_promotion").Collection("lucky_money"),
				collectionAmount: database_mgo.NewMongoDBconnection("mongodb://root:Vietnam2020@1.55.214.191:20003/").Database("service_promotion").Collection("lucky_money_amount"),
			},
			args:     args{},
			wantRes:  nil,
			wantErr:  false,
			lenGotGt: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &LixiCollection{
				collection:       tt.fields.collection,
				collectionAmount: tt.fields.collectionAmount,
			}
			got, err := o.FindStarted(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindStarted() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) < tt.lenGotGt {
				t.Error("FindStarted() error len")
			}

			for _, lixi := range got {
				for _, user := range lixi.UserIds {
					switch lixi.Method {
					case "RANDOM":
						fmt.Println(lixi, user)
					case "FIXED":
						fmt.Println(lixi, user)
					}
				}
			}

		})
	}
}

func TestLixiCollection_FindRandomAmount(t *testing.T) {
	type fields struct {
		collection       *mongo.Collection
		collectionAmount *mongo.Collection
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantRes  *entities.LixiAmount
		wantErr  bool
		lenGotGt int
	}{
		{
			name: "find all started",
			fields: fields{
				collection:       database_mgo.NewMongoDBconnection("mongodb://root:Vietnam2020@1.55.214.191:20003/").Database("service_promotion").Collection("lucky_money"),
				collectionAmount: database_mgo.NewMongoDBconnection("mongodb://root:Vietnam2020@1.55.214.191:20003/").Database("service_promotion").Collection("lucky_money_amount"),
			},
			args:     args{},
			wantRes:  nil,
			wantErr:  false,
			lenGotGt: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &LixiCollection{
				collection:       tt.fields.collection,
				collectionAmount: tt.fields.collectionAmount,
			}
			gotRes, err := o.FindRandomAmount(tt.args.ctx, 1000000, 2000000)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindRandomAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.lenGotGt > 0 && gotRes.ID == "" {
				t.Error("FindRandomAmount()")
			}
		})
	}
}
