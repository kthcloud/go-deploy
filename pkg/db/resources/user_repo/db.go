package user_repo

import (
	"context"
	"fmt"
	"go-deploy/models/model"
	"go-deploy/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

// Create creates a new user.
// If the user already exists, it will update the user
// as a way of synchronizing the user's information.
func (client *Client) Create(id string, params *model.UserCreateParams) (*model.User, error) {
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
		// update roles
		update := bson.D{{"$set", bson.D{
			{"username", params.Username},
			{"firstName", params.FirstName},
			{"lastName", params.LastName},
			{"email", params.Email},
			{"effectiveRole", params.EffectiveRole},
			{"isAdmin", params.IsAdmin},
			{"lastAuthenticatedAt", time.Now()},
		}}}
		err = client.UpdateWithBsonByID(id, update)
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
		EffectiveRole:       *params.EffectiveRole,
		PublicKeys:          []model.PublicKey{},
		LastAuthenticatedAt: time.Now(),

		Onboarded: false,
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
	db.AddIfNotNil(&updateData, "onboarded", params.Onboarded)

	if len(updateData) == 0 {
		return nil
	}

	err := client.SetWithBsonByID(id, updateData)
	if err != nil {
		return fmt.Errorf("failed to update user for %s. details: %w", id, err)
	}

	return nil
}
