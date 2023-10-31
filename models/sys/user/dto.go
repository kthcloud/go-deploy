package user

import (
	"github.com/fatih/structs"
	roleModel "go-deploy/models/config/role"
	"go-deploy/models/dto/body"
	"log"
)

func (user *User) ToDTO(effectiveRole *roleModel.Role, usage *Usage, storageURL *string) body.UserRead {
	publicKeys := make([]body.PublicKey, len(user.PublicKeys))
	for i, key := range user.PublicKeys {
		publicKeys[i] = body.PublicKey{
			Name: key.Name,
			Key:  key.Key,
		}
	}

	if usage == nil {
		usage = &Usage{}
	}

	if effectiveRole == nil {
		log.Println("effective role is nil when creating user read for user", user.Username)
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
		ID:         user.ID,
		Username:   user.Username,
		Email:      user.Email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		PublicKeys: publicKeys,
		Onboarded:  user.Onboarded,

		Role: body.Role{
			Name:        effectiveRole.Name,
			Description: effectiveRole.Description,
			Permissions: permissions,
		},
		Admin: user.IsAdmin,

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
