package storage_manager

import (
	"fmt"
	"go-deploy/models/dto/body"
	"go-deploy/service/deployment_service/constants"
)

func (storageManager *StorageManager) ToDTO() body.StorageManagerRead {
	var url *string
	ingress, ok := storageManager.Subsystems.K8s.IngressMap[constants.StorageManagerAppName]
	if ok && ingress.Created() && len(ingress.Hosts) > 0 && len(ingress.Hosts[0]) > 0 {
		fullURL := fmt.Sprintf("https://%s", ingress.Hosts[0])
		url = &fullURL
	}

	return body.StorageManagerRead{
		ID:        storageManager.ID,
		OwnerID:   storageManager.OwnerID,
		CreatedAt: storageManager.CreatedAt,
		Zone:      storageManager.Zone,
		URL:       url,
	}
}
