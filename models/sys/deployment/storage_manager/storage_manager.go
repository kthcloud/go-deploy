package storage_manager

import (
	"time"
)

type StorageManager struct {
	ID        string    `json:"id" bson:"id"`
	OwnerID   string    `json:"ownerId" bson:"ownerId"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`

	Subsystems Subsystems `json:"subsystems" bson:"subsystems"`

	Zone string `json:"zone" bson:"zone"`
}
