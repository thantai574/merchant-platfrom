package messages

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"orders-system/domain/entities"
	"orders-system/utils/helpers"
	"time"
)

type repoImpl struct {
	collectionMessages *mongo.Collection
}

func (r repoImpl) FindById(id string) (*entities.Message, error) {
	var message *entities.Message

	err := r.collectionMessages.FindOne(helpers.ContextWithTimeOut(), bson.M{"_id": id}).Decode(&message)

	return message, err
}

func (r repoImpl) ListMessage(offset int64, topics []string) ([]*entities.Message, error) {
	limit := int64(10)

	var messages []*entities.Message

	topics = append(topics, "/topics/wallet-all", "/topics/wallet-ios", "/topics/wallet-android")
	queryBson := bson.D{{Key: "topic", Value: bson.D{{Key: "$in", Value: topics}}}, {"status_push", 1}}

	cur, err := r.collectionMessages.Find(helpers.ContextWithTimeOut(), queryBson, &options.FindOptions{Skip: &offset, Limit: &limit, Sort: bson.D{
		{"updated_at", -1}},
	})

	if err != nil {
		return nil, err
	}
	defer cur.Close(helpers.ContextWithTimeOut())

	for cur.Next(helpers.ContextWithTimeOut()) {
		var message *entities.Message
		err = cur.Decode(&message)

		if message.TimePush != "" {
			message.CreatedAt, err = time.Parse("2006-01-02 15:04:05 -0700 MST", message.TimePush+" +0000 UTC")
			if err == nil {
				message.CreatedAt = message.CreatedAt.Add(-time.Hour * 7)
			}
		}

		messages = append(messages, message)
	}

	return messages, err
}

func (r repoImpl) CountMessageStatusUnread(uid string) (int64, error) {
	count, err := r.collectionMessages.CountDocuments(helpers.ContextWithTimeOut(), bson.D{{Key: "user_id", Value: uid}, {Key: "is_read", Value: false}})

	return count, err
}

func (r repoImpl) CreateMessage(message entities.Message) (*entities.Message, error) {
	_, err := r.collectionMessages.InsertOne(helpers.ContextWithTimeOut(), message)

	if err != nil {
		return nil, err
	}

	return r.FindById(message.Id)
}

func NewRepository(db *mongo.Client) *repoImpl {
	return &repoImpl{
		collectionMessages: db.Database("service_notification").Collection("messages"),
	}
}
