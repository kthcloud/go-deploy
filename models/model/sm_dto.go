package model

import (
	body "github.com/kthcloud/go-deploy/api/openapi/v2"
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
