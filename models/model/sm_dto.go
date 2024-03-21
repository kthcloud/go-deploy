package model

import (
	"go-deploy/dto/v1/body"
)

// ToDTO converts an SM to a body.SmRead DTO.
func (sm *SM) ToDTO() body.SmRead {
	return body.SmRead{
		ID:        sm.ID,
		OwnerID:   sm.OwnerID,
		CreatedAt: sm.CreatedAt,
		Zone:      sm.Zone,
		URL:       sm.GetURL(),
	}
}
