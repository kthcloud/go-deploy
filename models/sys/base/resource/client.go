package resource

import (
	"go-deploy/models/sys/base"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ResourceClient[T any] struct {
	Collection     *mongo.Collection
	IncludeDeleted bool
	Pagination     *base.Pagination
	ExtraFilter    *bson.D
}
