package banks

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"orders-system/domain/entities"
)

// RepoImpl -
type RepoImpl struct {
	collection *mongo.Collection
}

func (r RepoImpl) GetDetailLink(id string) (linked *entities.LinkedBankLink, err error) {
	result := &entities.LinkedBankLink{}
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = r.collection.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetUserLinkedList -
func (r RepoImpl) GetUserLinkedList(userID string) ([]*entities.LinkedBankLink, error) {
	result := []*entities.LinkedBankLink{}
	cur, err := r.collection.Find(context.TODO(),
		bson.M{"status": "success", "gpay_user_id": userID},
		&options.FindOptions{
			Sort: bson.M{"created_at": -1},
		})

	if err != nil {
		return nil, err
	}

	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		var link *entities.LinkedBankLink

		err = cur.Decode(&link)
		if err == nil {
			result = append(result, link)
		}

	}

	return result, nil
}

// NewLinkedBankLinkRepository -
func NewLinkedBankLinkRepository(db *mongo.Client, dbName string) *RepoImpl {
	return &RepoImpl{
		collection: db.Database(dbName).Collection("links"),
	}
}
