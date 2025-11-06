package devices

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"orders-system/domain/entities"
	"orders-system/utils/helpers"
)

type repoImpl struct {
	collection *mongo.Collection
}

func (r repoImpl) UpdateDevice(device_id, device_token string) error {
	filter := bson.D{{Key: "_id", Value: device_id}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "device_token", Value: device_token}}}}

	_, err := r.collection.UpdateOne(helpers.ContextWithTimeOut(), filter,
		update)
	return err
}

func (r repoImpl) CreateDevice(devices entities.Devices) (*entities.Devices, error) {
	_, err := r.collection.InsertOne(helpers.ContextWithTimeOut(), devices)

	if err != nil {
		return nil, err
	}

	return r.FindById(devices.Id)
}

func (r repoImpl) DeleteDevice(device_id, user_id string) error {

	_, err := r.collection.DeleteMany(helpers.ContextWithTimeOut(), bson.D{{
		"$or", bson.A{bson.D{{
			"user_id", user_id,
		}},
			bson.D{{
				"_id", device_id,
			}},
		},
	}})

	return err
}

func (r repoImpl) FindById(id string) (*entities.Devices, error) {
	var entity_device entities.Devices

	err := r.collection.FindOne(helpers.ContextWithTimeOut(), bson.M{"_id": id}).Decode(&entity_device)

	return &entity_device, err
}

func (r repoImpl) FindByUserId(user_id string) (*entities.Devices, error) {
	var entity_device entities.Devices

	err := r.collection.FindOne(helpers.ContextWithTimeOut(), bson.M{"user_id": user_id}).Decode(&entity_device)

	return &entity_device, err
}

func NewRepository(db *mongo.Client) *repoImpl {
	return &repoImpl{
		collection: db.Database("wallet").Collection("devices"),
	}
}
