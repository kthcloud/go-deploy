package user

import (
	"context"
	"fmt"
	"github.com/fatih/structs"
	"go-deploy/models"
	"go-deploy/models/dto/body"
	roleModel "go-deploy/models/sys/enviroment/role"
	"go.mongodb.org/mongo-driver/bson"
	"log"
)

func (u *User) ToDTO(effectiveRole *roleModel.Role, usage *Usage, storageURL *string) body.UserRead {
	publicKeys := make([]body.PublicKey, len(u.PublicKeys))
	for i, key := range u.PublicKeys {
		publicKeys[i] = body.PublicKey{
			Name: key.Name,
			Key:  key.Key,
		}
	}

	if usage == nil {
		usage = &Usage{}
	}

	if effectiveRole == nil {
		log.Println("effective role is nil when creating user read for user", u.Username)
		effectiveRole = &roleModel.Role{
			Name:        "unknown",
			Description: "unknown",
		}
	}

	permissionsStructMap := structs.Map(effectiveRole.Permissions)
	permissions := make([]string, 0)
	for name, value := range permissionsStructMap {
		hasPermission, ok := value.(bool)
		if ok && hasPermission {
			permissions = append(permissions, name)
		}
	}

	userRead := body.UserRead{
		ID:         u.ID,
		Username:   u.Username,
		Email:      u.Email,
		PublicKeys: publicKeys,
		Onboarded:  u.Onboarded,

		Role: body.Role{
			Name:        effectiveRole.Name,
			Description: effectiveRole.Description,
			Permissions: permissions,
		},
		Admin: u.IsAdmin,

		Quota: body.Quota{
			Deployments: effectiveRole.Quotas.Deployments,
			CpuCores:    effectiveRole.Quotas.CpuCores,
			RAM:         effectiveRole.Quotas.RAM,
			DiskSize:    effectiveRole.Quotas.DiskSize,
			Snapshots:   effectiveRole.Quotas.Snapshots,
		},
		Usage: body.Quota{
			Deployments: usage.Deployments,
			CpuCores:    usage.CpuCores,
			RAM:         usage.RAM,
			DiskSize:    usage.DiskSize,
			Snapshots:   usage.Snapshots,
		},

		StorageURL: storageURL,
	}

	return userRead
}

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
		_, err = models.UserCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return fmt.Errorf("failed to update user info for %s. details: %w", username, err)
		}

		return nil
	}

	_, err = models.UserCollection.InsertOne(context.TODO(), User{
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

func (client *Client) UpdateWithParams(id string, params *UserUpdate) error {
	updateData := bson.D{}

	models.AddIfNotNil(&updateData, "username", params.Username)
	models.AddIfNotNil(&updateData, "publicKeys", params.PublicKeys)
	models.AddIfNotNil(&updateData, "onboarded", params.Onboarded)

	if len(updateData) == 0 {
		return nil
	}

	err := client.UpdateWithBsonByID(id, updateData)
	if err != nil {
		return fmt.Errorf("failed to update user for %s. details: %w", id, err)
	}

	return nil
}
