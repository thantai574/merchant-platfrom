package database_mgo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func NewMongoDBconnection(uri string) *mongo.Client {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))

	if err != nil {
		panic(err.Error())
	}

	ctx, _ := context.WithTimeout(context.TODO(), time.Second*5)

	err = client.Connect(ctx)

	if err != nil {
		panic("can't be reachable server")
	}

	return client
}
