package model

import (
	"go-deploy/dto/v2/body"
)

// ToDTO converts an SM to a body.SmRead DTO.
func (sm *SM) ToDTO(externalPort *int) body.SmRead {
	return body.SmRead{
		ID:        sm.ID,
		OwnerID:   sm.OwnerID,
		CreatedAt: sm.CreatedAt,
		Zone:      sm.Zone,
		URL:       sm.GetURL(externalPort),
	}
}
