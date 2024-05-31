package base_clients

import (
	"context"
	"errors"
	"go-deploy/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// TimestampedResourceClient is a type of base client that adds methods to manage timestamped resources in the database.
// These resources are characterized by having data paired a timestamp, such as polled status
type TimestampedResourceClient[T any] struct {
	Collection   *mongo.Collection
	ExtraFilter  bson.M
	MaxDocuments int
}

// Save saves a resource to the database and delete old resources if the number of resources exceeds the limit
func (c *TimestampedResourceClient[T]) Save(resource *T) error {
	_, err := c.Collection.InsertOne(context.TODO(), resource)
	if err != nil {
		return err
	}

	err = c.deleteOldItems()
	if err != nil {
		return err
	}

	return nil
}

// List fetches resources from the database that are within the limit
func (c *TimestampedResourceClient[T]) List() ([]T, error) {
	return db.ListResources[T](c.Collection, db.GroupFilters(bson.D{}, c.ExtraFilter, nil, false), bson.D{}, &db.Pagination{Page: 0, PageSize: c.MaxDocuments}, &db.SortBy{Field: "timestamp", Order: -1})
}

func (c *TimestampedResourceClient[T]) deleteOldItems() error {
	// Fetch nth item
	skip := int64(c.MaxDocuments - 1)

	var withTimestamp struct {
		Timestamp time.Time `bson:"timestamp"`
	}

	err := c.Collection.FindOne(context.TODO(), bson.D{}, &options.FindOneOptions{
		Skip: &skip,
		Sort: bson.M{
			"timestamp": -1,
		},
	}).Decode(&withTimestamp)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil
		}

		return err
	}

	// Delete all items before nth item
	_, err = c.Collection.DeleteMany(context.TODO(), bson.M{
		"timestamp": bson.M{
			"$lt": withTimestamp.Timestamp,
		},
	})

	return err
}
