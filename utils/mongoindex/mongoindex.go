package mongoindex

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EnsureIndex will ensure the index model provided is on the given collection.
func EnsureIndex(ctx context.Context, c *mongo.Collection, keys []bson.E, unique bool) error {
	ks := bson.D{}
	indexNames := []string{}
	for _, k := range keys {
		indexNames = append(indexNames, fmt.Sprintf("%v_%v", k.Key, k.Value))
		ks = append(ks, k)
	}
	idxoptions := &options.IndexOptions{}
	idxoptions.SetBackground(true)
	idxoptions.SetUnique(unique)
	idm := mongo.IndexModel{
		Keys:    ks,
		Options: idxoptions,
	}

	idxs := c.Indexes()
	cur, err := idxs.List(ctx)
	if err != nil {
		return err
	}

	indexName := strings.Join(indexNames, "_")
	found := false
	for cur.Next(ctx) {
		d := bson.D{}
		cur.Decode(&d)

		for _, v := range d {
			if v.Key == "name" && v.Value == indexName {
				found = true
				break
			}
		}

	}

	if found {
		return nil
	}

	_, err = idxs.CreateOne(ctx, idm)
	if err != nil {
		fmt.Printf("create index error, name: %v, err: %v", indexName, err)
	}

	return err
}
