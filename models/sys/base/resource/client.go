package resource

import (
	"go-deploy/models"
	"go-deploy/models/sys/base"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Resource interface {
}

// ResourceClient is a type of base client that adds methods to manage a resource in the database.
// It includes many useful operations such as listing, searching, and pagination.
type ResourceClient[T Resource] struct {
	Collection     *mongo.Collection
	IncludeDeleted bool
	Pagination     *base.Pagination
	ExtraFilter    bson.M
	Search         *models.SearchParams
}

// AddExtraFilter adds an extra filter to the client.
func (client *ResourceClient[T]) AddExtraFilter(filter bson.D) *ResourceClient[T] {
	if client.ExtraFilter == nil {
		client.ExtraFilter = bson.M{
			"$and": bson.A{},
		}
	}

	client.ExtraFilter["$and"] = append(client.ExtraFilter["$and"].(bson.A), filter)

	return client
}
