package resource

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func (client *ResourceClient[T]) GetByID(id string) (*T, error) {
	return models.GetResource[T](client.Collection, bson.D{{"id", id}}, client.IncludeDeleted, client.ExtraFilter)
}

func (client *ResourceClient[T]) GetByName(name string) (*T, error) {
	return models.GetResource[T](client.Collection, bson.D{{"name", name}}, client.IncludeDeleted, client.ExtraFilter)
}

func (client *ResourceClient[T]) GetAll() ([]T, error) {
	return models.GetManyResources[T](client.Collection, bson.D{}, client.IncludeDeleted, client.Pagination, client.ExtraFilter)
}

func (client *ResourceClient[T]) ExistsByID(id string) (bool, error) {
	count, err := models.CountResources(client.Collection, bson.D{{"id", id}}, false, client.ExtraFilter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (client *ResourceClient[T]) UpdateWithBsonByID(id string, update bson.D) error {
	return models.UpdateOneResource(client.Collection, bson.D{{"id", id}}, update, client.IncludeDeleted, client.ExtraFilter)
}

func (client *ResourceClient[T]) UpdateWithBsonByName(name string, update bson.D) error {
	return models.UpdateOneResource(client.Collection, bson.D{{"name", name}}, update, client.IncludeDeleted, client.ExtraFilter)
}

func (client *ResourceClient[T]) SetWithBsonByID(id string, update bson.D) error {
	return client.UpdateWithBsonByID(id, bson.D{{"$set", update}})
}

func (client *ResourceClient[T]) SetWithBsonByName(name string, update bson.D) error {
	return client.UpdateWithBsonByName(name, bson.D{{"$set", update}})
}

func (client *ResourceClient[T]) DeleteByID(id string) error {
	_, err := client.Collection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{
			{"$set", bson.D{{"deletedAt", time.Now()}}},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to delete resource %s. details: %w", id, err)
	}

	return nil
}

func (client *ResourceClient[T]) Deleted(id string) (bool, error) {
	filter := bson.D{
		{"id", id},
		{"deletedAt", bson.M{"$nin": []interface{}{nil, time.Time{}}}},
	}
	count, err := models.CountResources(client.Collection, filter, true, client.ExtraFilter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (client *ResourceClient[T]) GetAllWithFilter(filter bson.D) ([]T, error) {
	return models.GetManyResources[T](client.Collection, filter, client.IncludeDeleted, client.Pagination, client.ExtraFilter)
}

func (client *ResourceClient[T]) CountWithFilter(filter bson.D) (int, error) {
	return models.CountResources(client.Collection, filter, client.IncludeDeleted, client.ExtraFilter)
}
