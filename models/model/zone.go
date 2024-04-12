package model

import "go-deploy/dto/v1/body"

const (
	// ZoneTypeDeployment is a zone type for deployments.
	// Deprecated: use capabilities instead
	ZoneTypeDeployment = "deployment"
	// ZoneTypeVM is a zone type for VMs.
	// Deprecated: use capabilities instead
	ZoneTypeVM = "vm"
)

type Zone struct {
	Name         string   `bson:"name"`
	Description  string   `bson:"description"`
	Capabilities []string `bson:"capabilities"`
	Interface    *string  `bson:"interface"`

	// Type
	// Deprecated: use capabilities instead
	Type string `bson:"type"`
}

// ToDTO converts a Zone to a body.ZoneRead DTO.
func (z *Zone) ToDTO() body.ZoneRead {
	return body.ZoneRead{
		Name:         z.Name,
		Description:  z.Description,
		Interface:    z.Interface,
		Capabilities: z.Capabilities,
		Type:         z.Type,
	}
}
