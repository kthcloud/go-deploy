package config

import "go-deploy/dto/v1/body"

// ToDTO converts a Zone to a body.ZoneRead DTO.
func (z *Zone) ToDTO() body.ZoneRead {
	capabilities := z.Capabilities
	if z.Capabilities == nil {
		capabilities = make([]string, 0)
	}

	var endpoints body.ZoneEndpoints
	if z.HasCapability(ZoneCapabilityDeployment) {
		endpoints.Deployment = z.Domains.ParentDeployment
		endpoints.Storage = z.Domains.ParentSM
	}

	if z.HasCapability(ZoneCapabilityVM) {
		endpoints.VM = z.Domains.ParentVM
		endpoints.VmApp = z.Domains.ParentVmApp
	}

	return body.ZoneRead{
		Name:         z.Name,
		Description:  z.Description,
		Capabilities: capabilities,
		Endpoints:    endpoints,
		Legacy:       false,
		Enabled:      z.Enabled,

		// Deprecated fields
		Interface: &z.Domains.ParentDeployment,
		Type:      "deployment",
	}
}

// ToDTO converts a LegacyZone to a body.ZoneRead DTO.
func (z *LegacyZone) ToDTO() body.ZoneRead {
	return body.ZoneRead{
		Name:         z.Name,
		Description:  z.Description,
		Capabilities: []string{ZoneCapabilityVM},
		Endpoints: body.ZoneEndpoints{
			VM: z.ParentDomain,
		},
		Legacy:  true,
		Enabled: false,

		// Deprecated fields
		Interface: &z.ParentDomain,
		Type:      "vm",
	}
}
