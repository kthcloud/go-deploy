package activityResource

import (
	"go-deploy/models/sys/base"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ActivityResourceClient[T any] struct {
	Collection  *mongo.Collection
	Pagination  *base.Pagination
	ExtraFilter bson.D
}
