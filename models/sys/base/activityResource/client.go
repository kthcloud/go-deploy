package activityResource

import (
	"go-deploy/models"
	"go-deploy/models/sys/base"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ActivityResourceClient[T any] struct {
	Collection  *mongo.Collection
	Pagination  *base.Pagination
	ExtraFilter bson.M
	Search      *models.SearchParams
}

func (client *ActivityResourceClient[T]) AddExtraFilter(filter bson.D) *ActivityResourceClient[T] {
	if client.ExtraFilter == nil {
		client.ExtraFilter = bson.M{
			"$and": bson.A{},
		}
	}

	client.ExtraFilter["$and"] = append(client.ExtraFilter["$and"].(bson.A), filter)

	return client
}
