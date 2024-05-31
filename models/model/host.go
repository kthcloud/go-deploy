package model

import (
	"fmt"
	"go-deploy/dto/v2/body"
	"time"
)

type Host struct {
	Name        string `json:"name" bson:"name"`
	DisplayName string `json:"displayName" bson:"displayName"`
	Zone        string `json:"zone" bson:"zone"`

	IP   string `json:"ip" bson:"ip"`
	Port int    `json:"port" bson:"port"`

	Enabled bool `json:"enabled" bson:"enabled"`

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
func NewHostByParams(params *body.HostRegisterParams) *Host {
	return &Host{
		Name:        params.Name,
		DisplayName: params.DisplayName,
		Zone:        params.Zone,
		IP:          params.IP,
		Port:        params.Port,
		Enabled:     true,

		// These are set to the current time to simplify database queries
		LastSeenAt:   time.Now(),
		RegisteredAt: time.Now(),
	}
}
