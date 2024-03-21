package model

import "go-deploy/dto/v1/body"

const (
	// ZoneTypeDeployment is a zone type for deployments.
	ZoneTypeDeployment = "deployment"
	// ZoneTypeVM is a zone type for VMs.
	ZoneTypeVM = "vm"
)

type Zone struct {
	Name        string  `bson:"name"`
	Description string  `bson:"description"`
	Type        string  `bson:"type"`
	Interface   *string `bson:"interface"`
}

// ToDTO converts a Zone to a body.ZoneRead DTO.
func (z *Zone) ToDTO() body.ZoneRead {
	return body.ZoneRead{
		Name:        z.Name,
		Description: z.Description,
		Type:        z.Type,
		Interface:   z.Interface,
	}
}
