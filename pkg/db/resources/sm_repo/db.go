package sm_repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	// ErrAlreadyExists is returned when a storage manager already exists for a user.
	ErrAlreadyExists = fmt.Errorf("storage manager already exists for user")
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

	err := client.CreateIfUnique(id, sm, bson.D{{Key: "ownerId", Value: ownerID}})
	if err != nil {
		if errors.Is(err, db.ErrUniqueConstraint) {
			return nil, ErrAlreadyExists
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
func (client *Client) GetURL(externalPort *int) (*string, error) {
	sm, err := client.GetWithFilterAndProjection(bson.D{}, bson.D{{Key: "ownerId", Value: 1}, {Key: "subsystems.k8s.ingressMap", Value: 1}})
	if err != nil {
		return nil, err
	}

	if sm == nil {
		return nil, nil
	}

	return sm.GetURL(externalPort), nil
}

// DeleteSubsystem deletes a subsystem from a storage manager.
// It prepends the key with `subsystems` and unsets it.
func (client *Client) DeleteSubsystem(id, key string) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.UpdateWithBsonByID(id, bson.D{{Key: "$unset", Value: bson.D{{Key: subsystemKey, Value: ""}}}})
}

// SetSubsystem sets a subsystem in a storage manager.
// It prepends the key with `subsystems` and sets it.
func (client *Client) SetSubsystem(id, key string, update interface{}) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.SetWithBsonByID(id, bson.D{{Key: subsystemKey, Value: update}})
}

// MarkRepaired marks a storage manager as repaired.
// It sets RepairedAt and unsets the repairing activity.
func (client *Client) MarkRepaired(id string) error {
	filter := bson.D{{Key: "id", Value: id}}
	update := bson.D{
		{Key: "$set", Value: bson.D{{Key: "repairedAt", Value: time.Now()}}},
		{Key: "$unset", Value: bson.D{{Key: "activities.repairing", Value: ""}}},
	}

	_, err := client.ResourceClient.Collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}
