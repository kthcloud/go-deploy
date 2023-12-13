package sm

import (
	"fmt"
	"go-deploy/models/dto/body"
	"go-deploy/service/constants"
)

func (sm *SM) ToDTO() body.SmRead {
	var url *string
	ingress, ok := sm.Subsystems.K8s.IngressMap[constants.SmAppName]
	if ok && ingress.Created() && len(ingress.Hosts) > 0 && len(ingress.Hosts[0]) > 0 {
		fullURL := fmt.Sprintf("https://%s", ingress.Hosts[0])
		url = &fullURL
	}

	return body.SmRead{
		ID:        sm.ID,
		OwnerID:   sm.OwnerID,
		CreatedAt: sm.CreatedAt,
		Zone:      sm.Zone,
		URL:       url,
	}
}
