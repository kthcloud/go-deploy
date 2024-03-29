package model

import (
	"go-deploy/dto/v1/body"
	"log"
)

// ToDTO converts a User to a body.UserRead DTO.
func (user *User) ToDTO(effectiveRole *Role, usage *UserUsage, storageURL *string) body.UserRead {
	publicKeys := make([]body.PublicKey, len(user.PublicKeys))
	for i, key := range user.PublicKeys {
		publicKeys[i] = body.PublicKey{
			Name: key.Name,
			Key:  key.Key,
		}
	}

	if usage == nil {
		usage = &UserUsage{}
	}

	if effectiveRole == nil {
		log.Println("effective role is nil when creating user read for user", user.Username)
		effectiveRole = &Role{
			Name:        "unknown",
			Description: "unknown",
		}
	}

	userRead := body.UserRead{
		ID:         user.ID,
		Username:   user.Username,
		Email:      user.Email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		PublicKeys: publicKeys,
		Onboarded:  user.Onboarded,

		Role:  effectiveRole.ToDTO(false),
		Admin: user.IsAdmin,

		Quota: effectiveRole.Quotas.ToDTO(),
		Usage: usage.ToDTO(),

		StorageURL: storageURL,
	}

	return userRead
}

// ToDTO converts a Usage to a body.Usage DTO.
func (usage *UserUsage) ToDTO() body.Usage {
	return body.Usage{
		Deployments: usage.Deployments,
		CpuCores:    usage.CpuCores,
		RAM:         usage.RAM,
		DiskSize:    usage.DiskSize,
		Snapshots:   usage.Snapshots,
	}
}
