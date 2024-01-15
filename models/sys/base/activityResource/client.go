package activityResource

import (
	"go-deploy/models"
	"go-deploy/models/sys/base"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ActivityResourceClient is a type of base client that adds methods to manage activities in the database.
type ActivityResourceClient[T any] struct {
	Collection  *mongo.Collection
	Pagination  *base.Pagination
	ExtraFilter bson.M
	Search      *models.SearchParams
}

// AddExtraFilter adds an extra filter to the client.
func (client *ActivityResourceClient[T]) AddExtraFilter(filter bson.D) *ActivityResourceClient[T] {
	if client.ExtraFilter == nil {
		client.ExtraFilter = bson.M{
			"$and": bson.A{},
		}
	}

	client.ExtraFilter["$and"] = append(client.ExtraFilter["$and"].(bson.A), filter)

	return client
}
