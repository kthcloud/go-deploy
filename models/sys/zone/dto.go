package zone

import "go-deploy/models/dto/body"

func (z *Zone) ToDTO() body.ZoneRead {
	return body.ZoneRead{
		Name:        z.Name,
		Description: z.Description,
		Type:        z.Type,
		Interface:   z.Interface,
	}
}
