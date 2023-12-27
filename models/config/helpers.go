package config

import (
	roleModels "go-deploy/models/sys/role"
)

func (e *ConfigType) GetRole(roleName string) *roleModels.Role {
	for _, role := range e.Roles {
		if role.Name == roleName {
			return &role
		}
	}

	return nil
}

func (e *ConfigType) GetRolesByIamGroups(iamGroups []string) []roleModels.Role {
	var roles []roleModels.Role

	for _, role := range e.Roles {
		for _, iamGroup := range iamGroups {
			if role.IamGroup == iamGroup {
				roles = append(roles, role)
			}
		}
	}

	return roles
}

func (d *Deployment) GetZone(name string) *DeploymentZone {
	for _, zone := range d.Zones {
		if zone.Name == name {
			return &zone
		}
	}
	return nil
}

func (v *VM) GetZone(name string) *VmZone {
	for _, zone := range v.Zones {
		if zone.Name == name {
			return &zone
		}
	}
	return nil
}

func (v *VM) GetZoneByID(id string) *VmZone {
	for _, zone := range v.Zones {
		if zone.ZoneID == id {
			return &zone
		}
	}
	return nil
}
