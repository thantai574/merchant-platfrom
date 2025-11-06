package wallet_config

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"orders-system/domain/request_params"
	"orders-system/domain/value_objects"
	"orders-system/utils/configs"
)

type RepoImpl struct {
	conf                            *configs.Config
	collectionMerchantServiceConfig *mongo.Collection
	collectionRefundConfig          *mongo.Collection
}

func (r RepoImpl) GetMerchantConfig(ctx context.Context, req request_params.GetMerchantConfigReq) (res value_objects.GetMerchantConfigRes, err error) {
	filter := bson.M{"merchant_ids": bson.M{"$in": []string{req.MerchantId}},
		"settings": bson.M{"$elemMatch": bson.M{"service_type": req.ServiceType,
			"trans_type": req.TransType,
			"status":     "ACTIVE",
		}},
	}

	err = r.collectionMerchantServiceConfig.FindOne(ctx, filter).Decode(&res)
	if err != nil {
		return value_objects.GetMerchantConfigRes{}, err
	}
	return res, err
}

func (r RepoImpl) GetRefundConfig(ctx context.Context, req request_params.GetRefundConfigReq) (res value_objects.GetRefundConfigRes, err error) {
	filter := bson.M{"trans_type": req.TransType,
		"settings": bson.M{"$elemMatch": bson.M{"status": "ACTIVE", "source_of_funds": bson.M{"$in": []string{req.SourceOfFund}}}},
	}

	if req.MerchantId != "" {
		filter["merchant_ids"] = bson.M{"$in": []string{req.MerchantId}}
	}

	err = r.collectionRefundConfig.FindOne(ctx, filter, &options.FindOneOptions{
		Sort: bson.M{"created_at": -1},
	}).Decode(&res)

	return
}

func NewServiceWalletConfigRepository(db *mongo.Client, dbName string, conf *configs.Config) *RepoImpl {
	return &RepoImpl{
		conf:                            conf,
		collectionMerchantServiceConfig: db.Database(dbName).Collection("merchant_service_configs"),
		collectionRefundConfig:          db.Database(dbName).Collection("refund_settings"),
	}
}
