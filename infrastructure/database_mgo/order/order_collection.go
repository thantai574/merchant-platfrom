package order

import (
	"context"
	"fmt"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	"orders-system/proto/order_system"
	"orders-system/utils/configs"
	"orders-system/utils/context_grpc"
	"orders-system/utils/helpers"
	"orders-system/utils/mongoindex"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const LenID = 10

type OrderCollection struct {
	conf                 *configs.Config
	collection           *mongo.Collection
	collection_increment *mongo.Collection
}

func (o *OrderCollection) FindByID(ctx context.Context, orderID string) (res *entities.OrderEntity, err error) {
	return
}

func (o *OrderCollection) GetOrderMerchant(ctx context.Context, req order_system.GetOrderByMerchantReq) (*order_system.GetOrderByMerchantRes, error) {
	panic("")
}

func (o *OrderCollection) GetOrderByRefundId(ctx context.Context, refundId string) (res *entities.OrderEntity, err error) {
	err = o.collection.FindOne(ctx, bson.M{"refund_transaction_id": refundId}).Decode(&res)
	return
}

func (o *OrderCollection) CheckLuckyMoney(ctx context.Context, user_id, lucky_money_id string) (order_entity *entities.OrderEntity, err error) {
	err = o.collection.FindOne(ctx, bson.D{
		{"user_id", user_id},
		{"lucky_money_id", lucky_money_id},
	}).Decode(&order_entity)

	if c, ok := ctx.(*context_grpc.CtxGrpc); ok {
		c.SetOrderId(order_entity.OrderID)
	}
	return
}

func (o *OrderCollection) FindByOrderID(ctx context.Context, orderID string) (res *entities.OrderEntity, err error) {
	if c, ok := ctx.(*context_grpc.CtxGrpc); ok {
		c.SetOrderId(orderID)
	}
	err = o.collection.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&res)

	return
}

func (o *OrderCollection) FindByMerchantIdAndRefId(ctx context.Context, merchantId, refId string) (res *entities.OrderEntity, err error) {
	err = o.collection.FindOne(ctx, bson.M{"subscribe_merchant_id": merchantId, "ref_id": refId}).Decode(&res)

	return
}

func (o *OrderCollection) GetExpiredOrder(ctx context.Context) (res []*entities.OrderEntity, err error) {
	var limit int64 = 10

	getListExpiredOrder, err := o.collection.Find(ctx, bson.M{"$or": []interface{}{
		bson.M{"status": order_system.OrderStatus_ORDER_PENDING}, bson.M{"status": order_system.OrderStatus_ORDER_PROCESSING}},
		"expired_at": bson.M{"$ne": nil, "$lte": helpers.GetCurrentTime()}, "order_type": bson.M{"$ne": constants.TRANSTYPE_WALLET_REFUND}},
		&options.FindOptions{
			Limit: &limit,
		})

	if err == nil {
		for getListExpiredOrder.Next(helpers.ContextWithTimeOut()) {
			var orderEntities entities.OrderEntity

			err = getListExpiredOrder.Decode(&orderEntities)
			if err != nil {
				continue
			}

			res = append(res, &orderEntities)
		}
	}

	return res, err
}

func (o *OrderCollection) Create(ctx context.Context, entity *entities.OrderEntity) (res *entities.OrderEntity, err error) {
	entity.OrderID, err = o.incrementID(ctx, o.conf.Prefix+"GPOS")
	if err != nil {
		return
	}

	if c, ok := ctx.(*context_grpc.CtxGrpc); ok {
		c.SetOrderId(entity.OrderID)
	}

	_, err = o.collection.InsertOne(ctx, entity)

	if err == nil {
		res = entity
	}

	return
}

func (o *OrderCollection) ProcessingOrderByID(ctx context.Context, order_entity *entities.OrderEntity) (res *entities.OrderEntity, err error) {
	update, err := o.collection.ReplaceOne(ctx, bson.D{{"order_id", order_entity.OrderID}, {"expired_at", bson.M{"$ne": nil, "$gte": helpers.GetCurrentTime()}}}, order_entity)

	if err != nil {
		return
	}

	if update.ModifiedCount == 0 {
		err = fmt.Errorf("Giao dịch đã hết hạn thanh toán")
	}
	res = order_entity

	if c, ok := ctx.(*context_grpc.CtxGrpc); ok {
		c.SetOrderId(order_entity.OrderID)
	}

	return
}

func (o *OrderCollection) ReplaceByID(ctx context.Context, order_entity *entities.OrderEntity) (res *entities.OrderEntity, err error) {

	update, err := o.collection.ReplaceOne(ctx, bson.M{"order_id": order_entity.OrderID}, order_entity)

	if err != nil {
		return
	}

	if update.ModifiedCount == 0 {
		err = fmt.Errorf("UpdateOrder order_id %v not found", order_entity.OrderID)
	}
	res = order_entity

	if c, ok := ctx.(*context_grpc.CtxGrpc); ok {
		c.SetOrderId(order_entity.OrderID)
	}

	return
}

func (r OrderCollection) incrementID(ctx context.Context, prefix string) (string, error) {
	after := options.After
	date := time.Now().Add(7 * time.Hour).Format("20060102")
	idGenerate := struct {
		Date string `bson:"date"`
		Id   int64  `bson:"id"`
		Len  int    `bson:"len"`
	}{}
	err := r.collection_increment.FindOneAndUpdate(ctx, bson.M{
		"date": date,
	}, bson.M{"$inc": bson.M{"id": 1}}, &options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}).Decode(&idGenerate)

	if err != nil {
		if err != mongo.ErrNoDocuments {
			return "", err
		}

		idGenerate.Id = 1
		idGenerate.Date = date
		idGenerate.Len = LenID
		_, err := r.collection_increment.InsertOne(ctx, idGenerate)
		if err != nil {
			return "", err
		}
	}
	id_string := fmt.Sprint(idGenerate.Id)

	gt := idGenerate.Len - len(id_string)

	for i := 0; i < gt; i++ {
		id_string = "0" + id_string
	}

	return prefix + date + id_string, nil
}

func NewOrderCollectionImpl(db *mongo.Client, conf *configs.Config) *OrderCollection {
	c := db.Database("service_order_system").Collection("orders")
	ci := db.Database("service_order_system").Collection("orders_id_increment")

	mongoindex.EnsureIndex(context.TODO(), c, []bson.E{
		{Key: "date", Value: -1},
	}, false)

	mongoindex.EnsureIndex(context.TODO(), c, []bson.E{
		{Key: "id", Value: -1},
	}, false)

	mongoindex.EnsureIndex(context.TODO(), ci, []bson.E{
		{Key: "date", Value: -1},
	}, true)

	mongoindex.EnsureIndex(context.TODO(), ci, []bson.E{
		{Key: "order_id", Value: -1},
	}, false)

	mongoindex.EnsureIndex(context.TODO(), ci, []bson.E{
		{Key: "expired_at", Value: -1},
	}, false)

	return &OrderCollection{
		conf:                 conf,
		collection:           c,
		collection_increment: ci,
	}
}
