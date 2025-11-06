package va_VCCB

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"orders-system/domain/entities"
	"orders-system/utils/configs"
	"orders-system/utils/helpers"
)

const MSBLenID = 7

type RepoImpl struct {
	conf                   *configs.Config
	collectionAccounts     *mongo.Collection
	collectionMSBIncrement *mongo.Collection
}

func (r RepoImpl) CreateVA(request entities.VirtualAccounts) (res entities.VirtualAccounts, err error) {
	if request.IsAutoIncrement && request.Provider == "MSB" {
		firstPrefix := r.conf.VaMsb.GpayIdentifier
		secondPrefix := request.MerchantIdentifier

		request.AccountNunmber, err = r.IncrementID(context.TODO(), firstPrefix+secondPrefix, request.Provider)
		if err != nil {
			return
		}
	}
	_, err = r.collectionAccounts.InsertOne(helpers.ContextWithTimeOut(), request)
	if err != nil {
		return entities.VirtualAccounts{}, err
	}

	return r.GetVAByAccountNumber(request.AccountNunmber)
}

func (r RepoImpl) DeleteVAAccount(accountNumber string) error {
	_, err := r.collectionAccounts.DeleteOne(context.Background(), bson.M{"account_number": accountNumber})
	return err
}

func (r RepoImpl) IncrementID(ctx context.Context, prefix, provider string) (string, error) {
	after := options.After
	idGenerate := struct {
		Id      int64  `bson:"id"`
		Len     int    `bson:"len"`
		Provide string `json:"provider"`
	}{}

	err := r.collectionMSBIncrement.FindOneAndUpdate(ctx, bson.M{}, bson.M{"$inc": bson.M{"id": 1}}, &options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}).Decode(&idGenerate)

	if err != nil {
		if err != mongo.ErrNoDocuments {
			return "", err
		}

		idGenerate.Id = 1
		idGenerate.Len = MSBLenID
		_, err := r.collectionMSBIncrement.InsertOne(ctx, idGenerate)
		if err != nil {
			return "", err
		}
	}
	idString := fmt.Sprint(idGenerate.Id)

	gt := idGenerate.Len - len(idString)

	for i := 0; i < gt; i++ {
		idString = "0" + idString
	}

	return prefix + idString, nil
}

func (r RepoImpl) GetVAByMapCondition(mapId, mapType, merchantId string) (res entities.VirtualAccounts, err error) {
	err = r.collectionAccounts.FindOne(helpers.ContextWithTimeOut(), bson.M{"map_id": mapId, "map_type": mapType, "merchant_id": merchantId, "status": "OPEN"}).Decode(&res)
	if err != nil {
		return entities.VirtualAccounts{}, err
	}
	return res, err
}

func (r RepoImpl) GetVAByMerchantId(merchantId string) (res entities.VirtualAccounts, err error) {
	err = r.collectionAccounts.FindOne(helpers.ContextWithTimeOut(), bson.M{"merchant_id": merchantId}).Decode(&res)
	if err != nil {
		return entities.VirtualAccounts{}, err
	}
	return res, err
}

func (r RepoImpl) GetVAByAccountNumber(accountNumber string) (res entities.VirtualAccounts, err error) {
	err = r.collectionAccounts.FindOne(helpers.ContextWithTimeOut(), bson.M{"account_number": accountNumber}).Decode(&res)
	if err != nil {
		return entities.VirtualAccounts{}, err
	}
	return res, err
}

func (r RepoImpl) IncrementBalanceVA(accountNumber string, balance int64) (res entities.VirtualAccounts, err error) {
	_, err = r.collectionAccounts.UpdateOne(helpers.ContextWithTimeOut(), bson.M{"account_number": accountNumber}, bson.M{
		"$inc": bson.M{
			"balance": balance,
		},
	})

	if err != nil {
		return entities.VirtualAccounts{}, err
	}

	return r.GetVAByAccountNumber(accountNumber)
}

func (r RepoImpl) UpdateVA(accountNumber string, fieldUpdate bson.M) (response entities.VirtualAccounts, err error) {
	_, err = r.collectionAccounts.UpdateOne(helpers.ContextWithTimeOut(), bson.M{"account_number": accountNumber}, bson.M{"$set": fieldUpdate})
	if err != nil {
		return entities.VirtualAccounts{}, err
	}

	return r.GetVAByAccountNumber(accountNumber)
}

func NewVAVCCBRepository(db *mongo.Client, dbName string, conf *configs.Config) *RepoImpl {
	return &RepoImpl{
		conf:                   conf,
		collectionAccounts:     db.Database(dbName).Collection("va_accounts"),
		collectionMSBIncrement: db.Database(dbName).Collection("va_msb_increment"),
	}
}
