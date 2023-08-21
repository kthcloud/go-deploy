package user

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go-deploy/models/dto/body"
	roleModel "go-deploy/models/sys/enviroment/role"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (u *User) ToDTO(quotas *roleModel.Quotas, usage *Usage) body.UserRead {
	publicKeys := make([]body.PublicKey, len(u.PublicKeys))
	for i, key := range u.PublicKeys {
		publicKeys[i] = body.PublicKey{
			Name: key.Name,
			Key:  key.Key,
		}
	}

	if quotas == nil {
		quotas = &roleModel.Quotas{}
	}

	if usage == nil {
		usage = &Usage{}
	}

	if u.Roles == nil {
		u.Roles = []string{}
	}

	userRead := body.UserRead{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Roles:    u.Roles,
		Quota: body.Quota{
			Deployments: quotas.Deployments,
			CpuCores:    quotas.CpuCores,
			RAM:         quotas.RAM,
			DiskSize:    quotas.DiskSize,
		},
		Usage: body.Quota{
			Deployments: usage.Deployments,
			CpuCores:    usage.CpuCores,
			RAM:         usage.RAM,
			DiskSize:    usage.DiskSize,
		},
		PublicKeys: publicKeys,
	}

	return userRead
}

func Create(id, username string, roles []string) error {
	current, err := GetByID(id)
	if err != nil {
		return err
	}

	if current != nil {
		// update roles
		filter := bson.D{{"id", id}}
		update := bson.D{{"$set", bson.D{
			{"roles", roles},
		}}}
		_, err = models.UserCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return fmt.Errorf("failed to update user info for %s. details: %s", username, err)
		}

		return nil
	}

	_, err = models.UserCollection.InsertOne(context.TODO(), User{
		ID:         id,
		Username:   username,
		Email:      "",
		Roles:      roles,
		PublicKeys: []PublicKey{},
	})

	if err != nil {
		return fmt.Errorf("failed to create user info for %s. details: %s", username, err)
	}

	return nil
}

func GetByID(id string) (*User, error) {
	var user User
	filter := bson.D{{"id", id}}
	err := models.UserCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		err = fmt.Errorf("failed to fetch user info by id %s. details: %s", id, err)
		return nil, err
	}

	return &user, err
}

func GetAll() ([]User, error) {
	var users []User
	cursor, err := models.UserCollection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}

	err = cursor.All(context.Background(), &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func Update(userID string, update *UserUpdate) error {
	updateData := bson.M{}

	models.AddIfNotNil(updateData, "username", update.Username)
	models.AddIfNotNil(updateData, "email", update.Email)
	models.AddIfNotNil(updateData, "publicKeys", update.PublicKeys)

	if len(updateData) == 0 {
		return nil
	}

	filter := bson.D{{"id", userID}}
	updateDoc := bson.D{{"$set", updateData}}

	_, err := models.UserCollection.UpdateOne(context.Background(), filter, updateDoc)
	if err != nil {
		return fmt.Errorf("failed to update user info for %s. details: %s", userID, err)
	}

	return nil
}
