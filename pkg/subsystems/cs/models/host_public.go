package models

import (
	"go-deploy/pkg/imp/cloudstack"
	"time"
)

type HostPublic struct {
	ID            string    `bson:"id"`
	Name          string    `bson:"name"`
	State         string    `bson:"state"`
	ResourceState string    `bson:"resourceState"`
	CreatedAt     time.Time `bson:"createdAt"`
}

func CreatePublicFromGet(host *cloudstack.Host) *HostPublic {
	return &HostPublic{
		ID:            host.Id,
		Name:          host.Name,
		State:         host.State,
		ResourceState: host.Resourcestate,
		CreatedAt:     formatCreatedAt(host.Created),
	}
}
