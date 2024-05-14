package model

import (
	"fmt"
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

	Activities map[string]Activity `bson:"activities"`

	Subsystems SmSubsystems `bson:"subsystems"`
}

// GetURL returns the URL of the storage manager
// If the K8s ingress does not exist, it will return nil, or if the ingress does not have a host, it will return nil.
func (sm *SM) GetURL() *string {
	ingress := sm.Subsystems.K8s.GetIngress(fmt.Sprintf("sm-%s", sm.OwnerID))
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

// DoingActivity returns true if the deployment is doing the given activity.
func (sm *SM) DoingActivity(activity string) bool {
	for _, a := range sm.Activities {
		if a.Name == activity {
			return true
		}
	}
	return false
}
