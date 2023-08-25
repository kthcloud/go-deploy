package storage_manager

import (
	"go-deploy/models/dto/body"
)

func (storageManager *StorageManager) ToDTO() body.StorageManagerRead {
	return body.StorageManagerRead{
		ID:        storageManager.ID,
		OwnerID:   storageManager.OwnerID,
		CreatedAt: storageManager.CreatedAt,
		Zone:      storageManager.Zone,
	}
}
