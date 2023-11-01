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
	ExtraFilter    bson.D
	Search         *models.SearchParams
}
