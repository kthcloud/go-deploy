package model

import (
	"github.com/fatih/structs"
	"go-deploy/dto/v1/body"
	"sort"
)

type Role struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	IamGroup    string      `yaml:"iamGroup"`
	Permissions Permissions `yaml:"permissions"`
	Quotas      Quotas      `yaml:"quotas"`
}

// ToDTO converts a Role to a body.Role DTO.
func (r *Role) ToDTO(includeQuota bool) body.Role {
	permissionsStructMap := structs.Map(r.Permissions)
	permissions := make([]string, 0)
	for name, value := range permissionsStructMap {
		hasPermission, ok := value.(bool)
		if ok && hasPermission {
			permissions = append(permissions, name)
		}
	}

	var quota *body.Quota
	if includeQuota {
		dto := r.Quotas.ToDTO()
		quota = &dto
	}

	sort.Slice(permissions, func(i, j int) bool {
		return permissions[i] < permissions[j]
	})

	return body.Role{
		Name:        r.Name,
		Description: r.Description,
		Permissions: permissions,
		Quota:       quota,
	}
}
