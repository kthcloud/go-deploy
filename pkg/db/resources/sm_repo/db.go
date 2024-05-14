package sm_repo

import (
	"context"
	"errors"
	"fmt"
	"go-deploy/models/model"
	"go-deploy/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

var (
	// AlreadyExistsErr is returned when a storage manager already exists for a user.
	AlreadyExistsErr = fmt.Errorf("storage manager already exists for user")
)

// Create creates a new storage manager.
func (client *Client) Create(id, ownerID string, params *model.SmCreateParams) (*model.SM, error) {
	sm := &model.SM{
		ID:         id,
		OwnerID:    ownerID,
		Zone:       params.Zone,
		CreatedAt:  time.Now(),
		RepairedAt: time.Time{},
		DeletedAt:  time.Time{},
		Activities: make(map[string]model.Activity),
		Subsystems: model.SmSubsystems{},
	}

	err := client.CreateIfUnique(id, sm, bson.D{{"ownerId", ownerID}})
	if err != nil {
		if errors.Is(err, db.UniqueConstraintErr) {
			return nil, AlreadyExistsErr
		} else {
			return nil, err
		}
	}

	fetched, err := client.GetByID(id)
	if err != nil {
		return nil, err
	}

	return fetched, nil
}

// GetURL returns the URL of the storage manager.
// It uses projection to only fetch the URL.
func (client *Client) GetURL() (*string, error) {
	sm, err := client.GetWithFilterAndProjection(bson.D{}, bson.D{{"ownerId", 1}, {"subsystems.k8s.ingressMap", 1}})
	if err != nil {
		return nil, err
	}

	if sm == nil {
		return nil, nil
	}

	return sm.GetURL(), nil
}

// DeleteSubsystem deletes a subsystem from a storage manager.
// It prepends the key with `subsystems` and unsets it.
func (client *Client) DeleteSubsystem(id, key string) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.UpdateWithBsonByID(id, bson.D{{"$unset", bson.D{{subsystemKey, ""}}}})
}

// SetSubsystem sets a subsystem in a storage manager.
// It prepends the key with `subsystems` and sets it.
func (client *Client) SetSubsystem(id, key string, update interface{}) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.SetWithBsonByID(id, bson.D{{subsystemKey, update}})
}

// MarkRepaired marks a storage manager as repaired.
// It sets RepairedAt and unsets the repairing activity.
func (client *Client) MarkRepaired(id string) error {
	filter := bson.D{{"id", id}}
	update := bson.D{
		{"$set", bson.D{{"repairedAt", time.Now()}}},
		{"$unset", bson.D{{"activities.repairing", ""}}},
	}

	_, err := client.ResourceClient.Collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}
