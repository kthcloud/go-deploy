package config

import (
	"github.com/kthcloud/go-deploy/models/model"
	"gopkg.in/yaml.v3"
	"os"
	"sync"
	"time"
)

var (
	RolesLock sync.RWMutex
	// ReloadRoleInterval is the interval in minutes to reload the roles from the configuration file.
	ReloadRoleInterval = 1
	LastRoleReload     = time.Time{}
)

// LoadRoles loads all roles from the configuration file.
func (c *ConfigType) LoadRoles() {
	fp := c.Filepath

	content, err := os.ReadFile(fp)
	if err != nil {
		return
	}

	var tmpConfig ConfigType
	err = yaml.Unmarshal(content, &tmpConfig)
	if err != nil {
		return
	}

	c.Roles = tmpConfig.Roles
}

// GetRoles returns all roles.
// It reloads the roles from the configuration file if stale.
// All roles are loaded locally by the configuration file
func (c *ConfigType) GetRoles() []model.Role {
	RolesLock.RLock()
	if time.Since(LastRoleReload) > time.Duration(ReloadRoleInterval)*time.Minute {
		RolesLock.RUnlock()
		RolesLock.Lock()
		defer RolesLock.Unlock()

		c.LoadRoles()
		LastRoleReload = time.Now()
	} else {
		defer RolesLock.RUnlock()
	}

	return c.Roles
}

// GetRole returns the role with the given name.
// If the role is not found, nil is returned.
// All roles are loaded locally by the configuration file
func (c *ConfigType) GetRole(roleName string) *model.Role {
	for _, role := range c.GetRoles() {
		if role.Name == roleName {
			return &role
		}
	}

	return nil
}

// GetRolesByIamGroups returns all roles with an IAM group matching.
// If no roles are found, an empty slice is returned.
// All roles are loaded locally by the configuration file
func (c *ConfigType) GetRolesByIamGroups(iamGroups []string) []model.Role {
	var roles []model.Role

	for _, role := range c.GetRoles() {
		for _, iamGroup := range iamGroups {
			if role.IamGroup == iamGroup {
				roles = append(roles, role)
			}
		}
	}

	return roles
}

// EnabledZones returns all zones that are enabled.
// If no zones are found, an empty slice is returned.
// All zones are loaded locally by the configuration file
func (c *ConfigType) EnabledZones() []Zone {
	var zones []Zone

	for _, zone := range c.Zones {
		if zone.Enabled {
			zones = append(zones, zone)
		}
	}

	return zones
}

// EnabledZoneNames returns the names of all zones that are enabled.
// If no zones are found, an empty slice is returned.
// All zones are loaded locally by the configuration file
func (c *ConfigType) EnabledZoneNames() []string {
	var names []string

	for _, zone := range c.EnabledZones() {
		names = append(names, zone.Name)
	}

	return names
}

// GetZone returns the Zone with the given name.
// If the zone is not found, nil is returned.
// All zones are loaded locally by the configuration file
func (c *ConfigType) GetZone(name string) *Zone {
	for _, zone := range c.Zones {
		if zone.Name == name {
			return &zone
		}
	}
	return nil
}

// HasCapability returns true if the deployment zone has the given capability.
// If the capability is not found, false is returned.
// All capabilities are loaded locally by the configuration file
func (z *Zone) HasCapability(capability string) bool {
	for _, c := range z.Capabilities {
		if c == capability {
			return true
		}
	}
	return false
}
