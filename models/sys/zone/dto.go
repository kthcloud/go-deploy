package zone

import "go-deploy/models/dto/body"

// ToDTO converts a Zone to a body.ZoneRead DTO.
func (z *Zone) ToDTO() body.ZoneRead {
	return body.ZoneRead{
		Name:        z.Name,
		Description: z.Description,
		Type:        z.Type,
		Interface:   z.Interface,
	}
}
