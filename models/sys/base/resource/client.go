package resource

import (
	"go-deploy/models"
	"go-deploy/models/sys/base"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Resource interface {
}

type ResourceClient[T Resource] struct {
	Collection     *mongo.Collection
	IncludeDeleted bool
	Pagination     *base.Pagination
	ExtraFilter    bson.M
	Search         *models.SearchParams
}

func (client *ResourceClient[T]) AddExtraFilter(filter bson.D) *ResourceClient[T] {
	if client.ExtraFilter == nil {
		client.ExtraFilter = bson.M{
			"$and": bson.A{},
		}
	}

	client.ExtraFilter["$and"] = append(client.ExtraFilter["$and"].(bson.A), filter)

	return client
}
