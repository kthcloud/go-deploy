package activityResource

import "go.mongodb.org/mongo-driver/mongo"

type ActivityResourceClient[T any] struct {
	Collection *mongo.Collection
}
