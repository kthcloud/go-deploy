package resource

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type ResourceClient[T any] struct {
	Collection     *mongo.Collection
	IncludeDeleted bool
}
