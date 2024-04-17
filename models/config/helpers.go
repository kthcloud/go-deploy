package config

import (
	"go-deploy/models/model"
)

// GetRole returns the role with the given name.
// If the role is not found, nil is returned.
// All roles are loaded locally by the configuration file
func (e *ConfigType) GetRole(roleName string) *model.Role {
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
func (e *ConfigType) GetRolesByIamGroups(iamGroups []string) []model.Role {
	var roles []model.Role

	for _, role := range e.Roles {
		for _, iamGroup := range iamGroups {
			if role.IamGroup == iamGroup {
				roles = append(roles, role)
			}
		}
	}

	return roles
}

// GetZone returns the Zone with the given name.
// If the zone is not found, nil is returned.
// All zones are loaded locally by the configuration file
func (d *ConfigType) GetZone(name string) *Zone {
	for _, zone := range d.Zones {
		if zone.Name == name {
			return &zone
		}
	}
	return nil
}

// HasCapability returns true if the deployment zone has the given capability.
// If the capability is not found, false is returned.
// All capabilities are loaded locally by the configuration file
func (d *Zone) HasCapability(capability string) bool {
	for _, c := range d.Capabilities {
		if c == capability {
			return true
		}
	}
	return false
}

// GetLegacyZone returns the VM zone with the given name.
// If the zone is not found, nil is returned.
// All zones are loaded locally by the configuration file
// Deprecated: Use ConfigType.GetZone instead
func (v *VM) GetLegacyZone(name string) *LegacyZone {
	for _, zone := range v.Zones {
		if zone.Name == name {
			return &zone
		}
	}
	return nil
}

// GetLegacyZoneByID returns the VM zone with the given ID.
// If the zone is not found, nil is returned.
// All zones are loaded locally by the configuration file
// Deprecated: Use ConfigType.GetZone instead
func (v *VM) GetLegacyZoneByID(id string) *LegacyZone {
	for _, zone := range v.Zones {
		if zone.ZoneID == id {
			return &zone
		}
	}
	return nil
}
