package config

import (
	roleModels "go-deploy/models/sys/role"
)

// GetRole returns the role with the given name.
// If the role is not found, nil is returned.
// All roles are loaded locally by the configuration file
func (e *ConfigType) GetRole(roleName string) *roleModels.Role {
	for _, role := range e.Roles {
		if role.Name == roleName {
			return &role
		}
	}

	return nil
}

// GetRolesByIamGroups returns all roles with an IAM group matching.
// If no roles are found, an empty slice is returned.
// All roles are loaded locally by the configuration file
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

// GetZone returns the Deployment zone with the given name.
// If the zone is not found, nil is returned.
// All zones are loaded locally by the configuration file
func (d *Deployment) GetZone(name string) *DeploymentZone {
	for _, zone := range d.Zones {
		if zone.Name == name {
			return &zone
		}
	}
	return nil
}

// GetZone returns the VM zone with the given name.
// If the zone is not found, nil is returned.
// All zones are loaded locally by the configuration file
func (v *VM) GetZone(name string) *VmZone {
	for _, zone := range v.Zones {
		if zone.Name == name {
			return &zone
		}
	}
	return nil
}

// GetZoneByID returns the VM zone with the given ID.
// If the zone is not found, nil is returned.
// All zones are loaded locally by the configuration file
func (v *VM) GetZoneByID(id string) *VmZone {
	for _, zone := range v.Zones {
		if zone.ZoneID == id {
			return &zone
		}
	}
	return nil
}
