package storage_manager

import (
	"fmt"
	"time"
)

type StorageManager struct {
	ID      string `json:"id" bson:"id"`
	OwnerID string `json:"ownerId" bson:"ownerId"`
	Zone    string `json:"zone" bson:"zone"`

	CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
	RepairedAt time.Time `json:"repairAt" bson:"repairAt"`

	Activities []string   `json:"activities" bson:"activities"`
	Subsystems Subsystems `json:"subsystems" bson:"subsystems"`
}

func (storageManager *StorageManager) GetURL() *string {
	ingress, ok := storageManager.Subsystems.K8s.IngressMap["oauth-proxy"]
	if !ok || !ingress.Created() {
		return nil
	}

	if len(ingress.Hosts) > 0 && len(ingress.Hosts[0]) > 0 {
		host := ingress.Hosts[0]
		url := fmt.Sprintf("https://%s", host)
		return &url
	}

	return nil
}
