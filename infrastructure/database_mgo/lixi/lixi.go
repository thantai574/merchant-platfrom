package lixi

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"orders-system/domain/entities"
	"time"
)

type LixiCollection struct {
	collection       *mongo.Collection
	collectionAmount *mongo.Collection
}

func (o *LixiCollection) FindStarted(ctx context.Context) (res []*entities.LixiEntity, err error) {
	res = []*entities.LixiEntity{}
	cursor, err := o.collection.Find(ctx, bson.D{{"status", "ACTIVE"}, {
		"start_date", bson.M{
			"$lte": time.Now().Unix(),
		},
	}})
	defer cursor.Close(ctx)
	if err != nil {
		return
	}

	for cursor.Next(ctx) {
		var detail *entities.LixiEntity
		err_decode := cursor.Decode(&detail)

		if err_decode == nil {
			res = append(res, detail)
		}
	}
	return

}

func (o *LixiCollection) UpdateOne(ctx context.Context, lixi *entities.LixiEntity) (lixi_response *entities.LixiEntity, err error) {
	_, err = o.collection.ReplaceOne(ctx, bson.M{"_id": lixi.ID}, lixi)

	if err != nil {
		return
	}

	return
}

func (o *LixiCollection) FindRandomAmount(ctx context.Context, min, max int64) (res *entities.LixiAmount, err error) {

	cur, err := o.collectionAmount.Aggregate(ctx, bson.A{bson.D{{
		"$match", bson.D{{
			"status", "ACTIVE",
		},
			{"amount", bson.M{"$gte": min}},
			{"amount", bson.M{"$lte": max}},
		},
	},
	},
		bson.D{
			{"$sample", bson.M{"size": 1}},
		},
	})
	defer cur.Close(ctx)

	if err != nil {
		return
	}

	for cur.Next(ctx) {

		err_decode := cur.Decode(&res)
		if err_decode == nil {
			return
		}
	}

	return

}

func NewLixiCollectionImpl(db *mongo.Client) *LixiCollection {
	c := db.Database("service_promotion").Collection("lucky_money")
	c_amount := db.Database("service_promotion").Collection("lucky_money_amount")

	return &LixiCollection{
		collection:       c,
		collectionAmount: c_amount,
	}
}
