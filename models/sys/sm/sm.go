package sm

import (
	"fmt"
	"go-deploy/models/sys/activity"
	"go-deploy/pkg/subsystems"
	"time"
)

type SM struct {
	ID      string `bson:"id"`
	OwnerID string `bson:"ownerId"`
	Zone    string `bson:"zone"`

	CreatedAt  time.Time `bson:"createdAt"`
	RepairedAt time.Time `bson:"repairedAt"`
	DeletedAt  time.Time `bson:"deletedAt"`

	Activities map[string]activity.Activity `bson:"activities"`
	Subsystems Subsystems                   `bson:"subsystems"`
}

func (sm *SM) GetURL() *string {
	ingress := sm.Subsystems.K8s.GetIngress("storage-manager")
	if subsystems.NotCreated(ingress) {
		return nil
	}

	if len(ingress.Hosts) > 0 && len(ingress.Hosts[0]) > 0 {
		host := ingress.Hosts[0]
		url := fmt.Sprintf("https://%s", host)
		return &url
	}

	return nil
}
