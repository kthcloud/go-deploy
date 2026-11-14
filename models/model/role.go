package model

import (
	"slices"

	"github.com/fatih/structs"
	body2 "github.com/kthcloud/go-deploy/api/openapi/v2"
)

type Role struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	IamGroup    string      `yaml:"iamGroup"`
	Permissions Permissions `yaml:"permissions"`
	Quotas      Quotas      `yaml:"quotas"`
}

// ToDTO converts a Role to a body.Role DTO.
func (r *Role) ToDTO(includeQuota bool) body2.Role {
	permissionsStructMap := structs.Map(r.Permissions)
	permissions := make([]string, 0)
	for name, value := range permissionsStructMap {
		hasPermission, ok := value.(bool)
		if ok && hasPermission {
			permissions = append(permissions, name)
		}
	}

	// If we dont explicitly disallow it they have the perm
	if r.Permissions.UseVms == nil {
		permissions = append(permissions, "useVms")
	}

	var quota *body2.Quota
	if includeQuota {
		dto := r.Quotas.ToDTO()
		quota = &dto
	}

	slices.Sort(permissions)

	return body2.Role{
		Name:        r.Name,
		Description: r.Description,
		Permissions: permissions,
		Quota:       quota,
	}
}
