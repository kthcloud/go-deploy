package user_data

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

func (client *Client) Create(id, data, userID string) (*UserData, error) {
	current, err := client.GetByID(id)
	if err != nil {
		return nil, err
	}

	if current != nil {
		// update roles
		err = client.SetWithBsonByID(id, bson.D{{"data", data}})
		if err != nil {
			return nil, fmt.Errorf("failed to update user data for %s. details: %w", id, err)
		}

		return client.GetByID(id)
	}

	_, err = client.Collection.InsertOne(context.TODO(), UserData{
		ID:        id,
		Data:      data,
		UserID:    userID,
		CreatedAt: time.Now(),
	})

	if err != nil {
		// If there is a race condition, update the user data instead
		if mongo.IsDuplicateKeyError(err) {

			err = client.SetWithBsonByID(id, bson.D{{"data", data}})
			if err != nil {
				return nil, fmt.Errorf("failed to update user data for %s. details: %w", id, err)
			}

			return client.GetByID(id)
		}

		return nil, fmt.Errorf("failed to create user info for %s. details: %w", id, err)
	}

	return client.GetByID(id)
}

func (client *Client) Update(id, data string) (*UserData, error) {
	err := client.SetWithBsonByID(id, bson.D{{"data", data}})
	if err != nil {
		return nil, fmt.Errorf("failed to update user data for %s. details: %w", id, err)
	}

	return client.GetByID(id)
}
