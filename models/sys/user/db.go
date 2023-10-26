package user

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go.mongodb.org/mongo-driver/bson"
)

func (client *Client) Create(id, username, email string, isAdmin bool, effectiveRole *EffectiveRole) error {
	current, err := client.GetByID(id)
	if err != nil {
		return err
	}

	if effectiveRole == nil {
		effectiveRole = &EffectiveRole{
			Name:        "default",
			Description: "Default role for new users",
		}
	}

	if current != nil {
		// update roles
		filter := bson.D{{"id", id}}
		update := bson.D{{"$set", bson.D{
			{"username", username},
			{"email", email},
			{"effectiveRole", effectiveRole},
			{"isAdmin", isAdmin},
		}}}
		_, err = client.Collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return fmt.Errorf("failed to update user info for %s. details: %w", username, err)
		}

		return nil
	}

	_, err = client.Collection.InsertOne(context.TODO(), User{
		ID:            id,
		Username:      username,
		Email:         email,
		EffectiveRole: *effectiveRole,
		IsAdmin:       isAdmin,
		PublicKeys:    []PublicKey{},
	})

	if err != nil {
		return fmt.Errorf("failed to create user info for %s. details: %w", username, err)
	}

	return nil
}

func (client *Client) UpdateWithParams(id string, params *UpdateParams) error {
	updateData := bson.D{}

	models.AddIfNotNil(&updateData, "username", params.Username)
	models.AddIfNotNil(&updateData, "publicKeys", params.PublicKeys)
	models.AddIfNotNil(&updateData, "onboarded", params.Onboarded)

	if len(updateData) == 0 {
		return nil
	}

	err := client.SetWithBsonByID(id, updateData)
	if err != nil {
		return fmt.Errorf("failed to update user for %s. details: %w", id, err)
	}

	return nil
}
