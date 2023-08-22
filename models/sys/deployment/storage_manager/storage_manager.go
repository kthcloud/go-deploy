package storage_manager

import (
	"go-deploy/models/sys/deployment/subsystems"
	"time"
)

type Subsystems struct {
	K8s subsystems.K8s `json:"k8s" bson:"k8s"`
}

type StorageManager struct {
	ID         string     `json:"id" bson:"id"`
	OwnerID    string     `json:"ownerId" bson:"ownerId"`
	CreatedAt  time.Time  `json:"createdAt" bson:"createdAt"`
	Zone       string     `json:"zone" bson:"zone"`
	Subsystems Subsystems `json:"subsystems" bson:"subsystems"`
}
