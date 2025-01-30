package model

import (
	"fmt"
	"time"

	"github.com/kthcloud/go-deploy/dto/v2/body"
)

type Host struct {
	Name        string `json:"name" bson:"name"`
	DisplayName string `json:"displayName" bson:"displayName"`
	Zone        string `json:"zone" bson:"zone"`

	IP   string `json:"ip" bson:"ip"`
	Port int    `json:"port" bson:"port"`

	Enabled bool `json:"enabled" bson:"enabled"`
	// Schedulable is used to determine if a host is available for scheduling.
	// If a host is not schedulable, it will not be used for scheduling and if it has any GPUs, they won't be synchronized.
	// It is merely a synchronized state with the Kubernetes cluster to respect the host's state, such as when cordoning a node.
	Schedulable bool `json:"schedulable" bson:"schedulable"`

	DeactivatedUntil *time.Time `json:"deactivatedUntil,omitempty" bson:"deactivatedUntil,omitempty"`
	LastSeenAt       time.Time  `json:"lastSeenAt" bson:"lastSeenAt"`
	RegisteredAt     time.Time  `json:"registeredAt" bson:"registeredAt"`
}

func (host *Host) ApiURL() string {
	return fmt.Sprintf("http://%s:%d", host.IP, host.Port)
}

func (host *Host) ToDTO() body.HostRead {
	return body.HostRead{
		HostBase: body.HostBase{
			Name:        host.Name,
			DisplayName: host.DisplayName,
			Zone:        host.Zone,
		},
	}
}
func (host *Host) ToVerboseDTO() body.HostVerboseRead {
	return body.HostVerboseRead{
		HostBase: body.HostBase{
			Name:        host.Name,
			DisplayName: host.DisplayName,
			Zone:        host.Zone,
		},
		IP:               host.IP,
		Port:             host.Port,
		Enabled:          host.Enabled,
		Schedulable:      host.Schedulable,
		DeactivatedUntil: host.DeactivatedUntil,
		LastSeenAt:       host.LastSeenAt,
		RegisteredAt:     host.RegisteredAt,
	}
}
func NewHostByParams(params *body.HostRegisterParams) *Host {
	return &Host{
		Name:        params.Name,
		DisplayName: params.DisplayName,
		Zone:        params.Zone,
		IP:          params.IP,
		Port:        params.Port,

		Enabled:     params.Enabled,
		Schedulable: params.Schedulable,

		// These are set to the current time to simplify database queries
		LastSeenAt:   time.Now(),
		RegisteredAt: time.Now(),
	}
}
