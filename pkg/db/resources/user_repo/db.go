package user_repo

import (
	"context"
	"fmt"
	"go-deploy/models/model"
	"go-deploy/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

// Synchronize creates a new user or updates an existing user.
func (client *Client) Synchronize(id string, params *model.UserSynchronizeParams) (*model.User, error) {
	current, err := client.GetByID(id)
	if err != nil {
		return nil, err
	}

	if params.EffectiveRole == nil {
		params.EffectiveRole = &model.EffectiveRole{
			Name:        "default",
			Description: "Default role for new users",
		}
	}

	if current != nil {
		// Update the user
		update := bson.D{
			{"username", params.Username},
			{"firstName", params.FirstName},
			{"lastName", params.LastName},
			{"email", params.Email},
			{"effectiveRole", params.EffectiveRole},
			{"isAdmin", params.IsAdmin},
			{"lastAuthenticatedAt", time.Now()},
		}

		err = client.SetWithBsonByID(id, update)
		if err != nil {
			return nil, fmt.Errorf("failed to update user info for %s. details: %w", id, err)
		}

		return client.GetByID(id)
	}

	_, err = client.Collection.InsertOne(context.TODO(), model.User{
		ID:                  id,
		Username:            params.Username,
		FirstName:           params.FirstName,
		LastName:            params.LastName,
		Email:               params.Email,
		IsAdmin:             params.IsAdmin,
		Gravatar:            model.CreateEmptyGravatar(),
		EffectiveRole:       *params.EffectiveRole,
		PublicKeys:          []model.PublicKey{},
		LastAuthenticatedAt: time.Now(),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create user info for %s. details: %w", id, err)
	}

	return client.GetByID(id)
}

// UpdateWithParams updates the user with the given params.
func (client *Client) UpdateWithParams(id string, params *model.UserUpdateParams) error {
	updateData := bson.D{}

	db.AddIfNotNil(&updateData, "publicKeys", params.PublicKeys)
	db.AddIfNotNil(&updateData, "apiKeys", params.ApiKeys)
	db.AddIfNotNil(&updateData, "userData", params.UserData)

	if len(updateData) == 0 {
		return nil
	}

	err := client.SetWithBsonByID(id, updateData)
	if err != nil {
		return fmt.Errorf("failed to update user for %s. details: %w", id, err)
	}

	return nil
}

// SetGravatar updates the gravatar URL for the user.
func (client *Client) SetGravatar(id string, url string) error {
	update := bson.D{
		{"gravatar.url", url},
		{"gravatar.fetchedAt", time.Now()},
	}

	err := client.SetWithBsonByID(id, update)
	if err != nil {
		return fmt.Errorf("failed to update gravatar for %s. details: %w", id, err)
	}

	return nil
}

// UnsetGravatar removes the gravatar URL for the user.
func (client *Client) UnsetGravatar(id string) error {
	update := bson.D{
		{"$unset", bson.D{{"gravatar.url", ""}}},
		{"$set", bson.D{{"gravatar.fetchedAt", time.Now()}}},
	}

	err := client.UpdateWithBsonByID(id, update)
	if err != nil {
		return fmt.Errorf("failed to unset gravatar for %s. details: %w", id, err)
	}

	return nil
}
