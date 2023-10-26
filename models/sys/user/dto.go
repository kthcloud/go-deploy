package user

import (
	"github.com/fatih/structs"
	"go-deploy/models/dto/body"
	roleModel "go-deploy/models/sys/enviroment/role"
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
