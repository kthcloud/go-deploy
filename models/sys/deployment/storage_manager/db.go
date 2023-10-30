package storage_manager

import (
	"fmt"
	"go-deploy/models"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

var AlreadyExistsErr = fmt.Errorf("storage manager already exists for user")

func (client *Client) CreateStorageManager(id string, params *CreateParams) (*StorageManager, error) {
	storageManager := &StorageManager{
		ID:        id,
		OwnerID:   params.UserID,
		CreatedAt: time.Now(),
		Zone:      params.Zone,
	}

	err := client.CreateIfUnique(id, storageManager, bson.D{{"ownerId", params.UserID}})
	if err != nil {
		if err == models.UniqueConstraintErr {
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
