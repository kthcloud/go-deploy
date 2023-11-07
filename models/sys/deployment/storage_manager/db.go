package storage_manager

import (
	"errors"
	"fmt"
	"go-deploy/models"
	"go-deploy/models/sys/activity"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

var AlreadyExistsErr = fmt.Errorf("storage manager already exists for user")

func (client *Client) CreateStorageManager(id string, params *CreateParams) (*StorageManager, error) {
	storageManager := &StorageManager{
		ID:         id,
		OwnerID:    params.UserID,
		Zone:       params.Zone,
		CreatedAt:  time.Now(),
		RepairedAt: time.Time{},
		Activities: map[string]activity.Activity{ActivityBeingCreated: {ActivityBeingCreated, time.Now()}},
		Subsystems: Subsystems{},
	}

	err := client.CreateIfUnique(id, storageManager, bson.D{{"ownerId", params.UserID}})
	if err != nil {
		if errors.Is(err, models.UniqueConstraintErr) {
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

func (client *Client) UpdateSubsystemByID(id, key string, update interface{}) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.SetWithBsonByID(id, bson.D{{subsystemKey, update}})
}
