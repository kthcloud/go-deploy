package base_clients

import (
	"fmt"
	"github.com/kthcloud/go-deploy/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type Resource interface {
}

// ResourceClient is a type of base client that adds methods to manage a model in the database.
// It includes many useful operations such as listing, searching, and pagination.
type ResourceClient[T Resource] struct {
	Collection     *mongo.Collection
	IncludeDeleted bool
	ExtraFilter    bson.M
	Pagination     *db.Pagination
	Search         *db.SearchParams
	SortBy         *db.SortBy
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

// GetByID returns a model with the given ID.
func (client *ResourceClient[T]) GetByID(id string) (*T, error) {
	return db.GetResource[T](client.Collection, db.GroupFilters(bson.D{{"id", id}}, client.ExtraFilter, client.Search, client.IncludeDeleted), nil)
}

// GetByName returns a model with the given name.
func (client *ResourceClient[T]) GetByName(name string) (*T, error) {
	return db.GetResource[T](client.Collection, db.GroupFilters(bson.D{{"name", name}}, client.ExtraFilter, client.Search, client.IncludeDeleted), nil)
}

// List returns any resources that match the given filter.
func (client *ResourceClient[T]) List() ([]T, error) {
	return db.ListResources[T](client.Collection, db.GroupFilters(bson.D{}, client.ExtraFilter, client.Search, client.IncludeDeleted), nil, client.Pagination, client.SortBy)
}

// ExistsByID returns whether a model with the given ID exists.
func (client *ResourceClient[T]) ExistsByID(id string) (bool, error) {
	count, err := db.CountResources(client.Collection, db.GroupFilters(bson.D{{"id", id}}, client.ExtraFilter, client.Search, client.IncludeDeleted))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsByName returns whether a model with the given name exists.
func (client *ResourceClient[T]) ExistsByName(name string) (bool, error) {
	count, err := db.CountResources(client.Collection, db.GroupFilters(bson.D{{"name", name}}, client.ExtraFilter, client.Search, client.IncludeDeleted))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsWithFilter returns whether a model with the given filter exists.
func (client *ResourceClient[T]) ExistsWithFilter(filter bson.D) (bool, error) {
	count, err := db.CountResources(client.Collection, db.GroupFilters(filter, client.ExtraFilter, client.Search, client.IncludeDeleted))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsAny returns whether any resources exist with the given filter.
func (client *ResourceClient[T]) ExistsAny() (bool, error) {
	count, err := db.CountResources(client.Collection, db.GroupFilters(bson.D{}, client.ExtraFilter, client.Search, client.IncludeDeleted))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CreateIfUnique creates a model with the given ID and model,
// but only if a model with the given filter does not already exist.
func (client *ResourceClient[T]) CreateIfUnique(id string, resource *T, filter bson.D) error {
	return db.CreateIfUniqueResource[T](client.Collection, id, resource, db.GroupFilters(filter, client.ExtraFilter, client.Search, client.IncludeDeleted))
}

// UpdateWithBSON updates a model with the given BSON update.
func (client *ResourceClient[T]) UpdateWithBSON(update bson.D) error {
	return client.UpdateWithBsonByFilter(bson.D{}, update)
}

// UpdateWithBsonByID updates a model with the given ID and BSON update.
func (client *ResourceClient[T]) UpdateWithBsonByID(id string, update bson.D) error {
	return client.UpdateWithBsonByFilter(bson.D{{"id", id}}, update)
}

// UpdateWithBsonByName updates a model with the given name and BSON update.
func (client *ResourceClient[T]) UpdateWithBsonByName(name string, update bson.D) error {
	return client.UpdateWithBsonByFilter(bson.D{{"name", name}}, update)
}

// UpdateWithBsonByFilter updates a model with the given filter and BSON update.
func (client *ResourceClient[T]) UpdateWithBsonByFilter(filter bson.D, update bson.D) error {
	return db.UpdateOneResource(client.Collection, db.GroupFilters(filter, client.ExtraFilter, client.Search, client.IncludeDeleted), update)
}

// UnsetWithBSON unsets the given fields from a model.
func (client *ResourceClient[T]) UnsetWithBSON(fields ...string) error {
	update := bson.D{
		{"$unset", bson.D{}},
	}

	for _, field := range fields {
		update[0].Value = append(update[0].Value.(bson.D), bson.E{Key: field, Value: ""})
	}

	return client.UpdateWithBSON(update)
}

// UnsetByID unsets the given fields from a model with the given ID.
func (client *ResourceClient[T]) UnsetByID(id string, fields ...string) error {
	update := bson.D{
		{"$unset", bson.D{}},
	}

	for _, field := range fields {
		update[0].Value = append(update[0].Value.(bson.D), bson.E{Key: field, Value: ""})
	}

	return client.UpdateWithBsonByID(id, update)
}

// UnsetByName unsets the given fields from a model with the given name.
func (client *ResourceClient[T]) UnsetByName(name string, fields ...string) error {
	update := bson.D{
		{"$unset", bson.D{}},
	}

	for _, field := range fields {
		update[0].Value = append(update[0].Value.(bson.D), bson.E{Key: field, Value: ""})
	}

	return client.UpdateWithBsonByName(name, update)
}

// SetWithBSON sets the given fields to the given values in a model.
func (client *ResourceClient[T]) SetWithBSON(update bson.D) error {
	return client.UpdateWithBSON(bson.D{{"$set", update}})
}

// SetWithBsonByID sets the given fields to the given values in a model with the given ID.
func (client *ResourceClient[T]) SetWithBsonByID(id string, update bson.D) error {
	return client.UpdateWithBsonByID(id, bson.D{{"$set", update}})
}

// SetWithBsonByName sets the given fields to the given values in a model with the given name.
func (client *ResourceClient[T]) SetWithBsonByName(name string, update bson.D) error {
	return client.UpdateWithBsonByName(name, bson.D{{"$set", update}})
}

// SetWithBsonByFilter sets the given fields to the given values in a model with the given filter.
func (client *ResourceClient[T]) SetWithBsonByFilter(filter bson.D, update bson.D) error {
	return client.UpdateWithBsonByFilter(filter, bson.D{{"$set", update}})
}

// CountDistinct returns the number of distinct values for the given field.
func (client *ResourceClient[T]) CountDistinct(field string) (int, error) {
	return db.CountDistinctResources(client.Collection, field, db.GroupFilters(bson.D{}, client.ExtraFilter, client.Search, client.IncludeDeleted))
}

// Delete deletes all resources that match the given filter.
// It only sets the deletedAt field to the current time (which
// will cause it be to be treated as a deleted model), and
// does not remove the resources from the database.
func (client *ResourceClient[T]) Delete() error {
	update := bson.D{
		{"$set", bson.D{{"deletedAt", time.Now()}}},
	}

	err := client.UpdateWithBSON(update)
	if err != nil {
		return fmt.Errorf("failed to delete resources. details: %w", err)
	}

	return nil
}

// DeleteByID deletes a model with the given ID.
// It only sets the deletedAt field to the current time (which
// // will cause it be to be treated as a deleted model), and
// does not remove the model from the database.
func (client *ResourceClient[T]) DeleteByID(id string) error {
	update := bson.D{
		{"$set", bson.D{{"deletedAt", time.Now()}}},
	}

	err := client.UpdateWithBsonByID(id, update)
	if err != nil {
		return fmt.Errorf("failed to delete model. details: %w", err)
	}

	return nil
}

// Deleted returns whether a model with the given ID has been deleted.
func (client *ResourceClient[T]) Deleted(id string) (bool, error) {
	filter := bson.D{
		{"id", id},
		{"deletedAt", bson.M{"$nin": []interface{}{nil, time.Time{}}}},
	}
	count, err := db.CountResources(client.Collection, db.GroupFilters(filter, client.ExtraFilter, client.Search, true))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Erase deletes all resources that match the given filter.
// It removes the resources from the database.
func (client *ResourceClient[T]) Erase() error {
	return db.DeleteResources(client.Collection, db.GroupFilters(bson.D{}, client.ExtraFilter, client.Search, client.IncludeDeleted))
}

// EraseByID deletes a model with the given ID.
// It removes the model from the database.
func (client *ResourceClient[T]) EraseByID(id string) error {
	return db.DeleteResources(client.Collection, db.GroupFilters(bson.D{{"id", id}}, client.ExtraFilter, client.Search, client.IncludeDeleted))
}

// Get returns a model with the given filter.
func (client *ResourceClient[T]) Get() (*T, error) {
	return db.GetResource[T](client.Collection, db.GroupFilters(bson.D{}, client.ExtraFilter, client.Search, client.IncludeDeleted), nil)
}

// OnlyID is a type that only contains an ID.
// This is useful when only the ID is needed.
// This should be paired with a projection that only includes the ID,
// such as bson.D{{"id", 1}}.
type OnlyID struct {
	ID string `bson:"id"`
}

// OnlyName is a type that only contains a name.
// This is useful when only the name is needed.
// This should be paired with a projection that only includes the name,
// such as bson.D{{"name", 1}}.
type OnlyName struct {
	Name string `bson:"name"`
}

// GetID returns the ID of a model with the given filter.
// It returns a OnlyID type, which only contains the ID.
func (client *ResourceClient[T]) GetID() (*string, error) {
	projection := bson.D{{"id", 1}}

	idStruct, err := db.GetResource[OnlyID](client.Collection, db.GroupFilters(bson.D{}, client.ExtraFilter, client.Search, client.IncludeDeleted), projection)
	if err != nil {
		return nil, err
	}

	if idStruct == nil {
		return nil, nil
	}

	return &idStruct.ID, nil
}

// GetName returns the name of a model with the given filter.
// It returns a OnlyName type, which only contains the name.
func (client *ResourceClient[T]) GetName(id string) (*string, error) {
	projection := bson.D{{"name", 1}}

	nameStruct, err := db.GetResource[OnlyName](client.Collection, db.GroupFilters(bson.D{{"id", id}}, client.ExtraFilter, client.Search, client.IncludeDeleted), projection)
	if err != nil {
		return nil, err
	}

	if nameStruct == nil {
		return nil, nil
	}

	return &nameStruct.Name, nil
}

// ListIDs returns the IDs of all resources that match the given filter.
func (client *ResourceClient[T]) ListIDs() ([]string, error) {
	projection := bson.D{{"id", 1}}
	ids, err := db.ListResources[OnlyID](client.Collection, db.GroupFilters(nil, client.ExtraFilter, client.Search, client.IncludeDeleted), projection, client.Pagination, client.SortBy)
	if err != nil {
		return nil, err
	}

	idList := make([]string, len(ids))
	for i, id := range ids {
		idList[i] = id.ID
	}

	return idList, nil
}

// ListNames returns the names of all resources that match the given filter.
func (client *ResourceClient[T]) ListNames() ([]string, error) {
	projection := bson.D{{"name", 1}}
	names, err := db.ListResources[OnlyName](client.Collection, db.GroupFilters(nil, client.ExtraFilter, client.Search, client.IncludeDeleted), projection, client.Pagination, client.SortBy)
	if err != nil {
		return nil, err
	}

	nameList := make([]string, len(names))
	for i, name := range names {
		nameList[i] = name.Name
	}

	return nameList, nil
}

// GetWithFilterAndProjection returns a model with the given filter and projection.
// It should be used as a base method for other methods to use, and not called directly.
func (client *ResourceClient[T]) GetWithFilterAndProjection(filter, projection bson.D) (*T, error) {
	return db.GetResource[T](client.Collection, db.GroupFilters(filter, client.ExtraFilter, client.Search, client.IncludeDeleted), projection)
}

// ListWithFilterAndProjection returns any resources that match the given filter and projection.
// It should be used as a base method for other methods to use, and not called directly.
func (client *ResourceClient[T]) ListWithFilterAndProjection(filter, projection bson.D) ([]T, error) {
	return db.ListResources[T](client.Collection, db.GroupFilters(filter, client.ExtraFilter, client.Search, client.IncludeDeleted), projection, client.Pagination, client.SortBy)
}

// Count returns the number of resources that match the given filter.
func (client *ResourceClient[T]) Count() (int, error) {
	return db.CountResources(client.Collection, db.GroupFilters(bson.D{}, client.ExtraFilter, client.Search, client.IncludeDeleted))
}
