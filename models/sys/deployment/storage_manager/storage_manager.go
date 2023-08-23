package storage_manager

import "time"

type StorageManager struct {
	ID      string `json:"id" bson:"id"`
	OwnerID string `json:"ownerId" bson:"ownerId"`
	Zone    string `json:"zone" bson:"zone"`

	CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
	RepairedAt time.Time `json:"repairAt" bson:"repairAt"`

	Activities []string   `json:"activities" bson:"activities"`
	Subsystems Subsystems `json:"subsystems" bson:"subsystems"`
}
