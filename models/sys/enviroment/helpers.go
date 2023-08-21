package enviroment

import roleModel "go-deploy/models/sys/enviroment/role"

func (e *Environment) GetRole(roleName string) *roleModel.Role {
	for _, role := range e.Roles {
		if role.Name == roleName {
			return &role
		}
	}

	return nil
}

func (e *Environment) GetRolesByIamGroups(iamGroups []string) []roleModel.Role {
	var roles []roleModel.Role

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
