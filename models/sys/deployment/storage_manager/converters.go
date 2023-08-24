package storage_manager

import (
	"go-deploy/models/dto/body"
	"time"
)

func (storageManager *StorageManager) ToDTO() body.StorageManagerRead {
	return body.StorageManagerRead{
		ID:        storageManager.ID,
		OwnerID:   storageManager.OwnerID,
		CreatedAt: time.Now(),
		Zone:      storageManager.Zone,
	}
}
