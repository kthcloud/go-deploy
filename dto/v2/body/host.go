package body

import "time"

type HostRead struct {
	HostBase `json:",inline" tstype:",extends"`
}

type HostVerboseRead struct {
	HostBase         `json:",inline" tstype:",extends"`
	IP               string     `json:"ip"`
	Port             int        `json:"port"`
	Enabled          bool       `json:"enabled"`
	Schedulable      bool       `json:"schedulable"`
	DeactivatedUntil *time.Time `json:"deactivatedUntil,omitempty"`
	LastSeenAt       time.Time  `json:"lastSeenAt"`
	RegisteredAt     time.Time  `json:"registeredAt"`
}

type HostBase struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	// Zone is the name of the zone where the host is located.
	Zone string `json:"zone"`
}

type HostRegisterParams struct {
	// Name is the host name of the node
	Name string `json:"name" binding:"required,min=3"`
	// DisplayName is the human-readable name of the node
	// This is optional, and is set to Name if not provided
	DisplayName string `json:"displayName" binding:"min=3"`
	IP          string `json:"ip" binding:"required"`
	// Port is the port the node is listening on for API requests
	Port int    `json:"port" binding:"required"`
	Zone string `json:"zone" binding:"required"`

	// Token is the discovery token validated against the config
	Token string `json:"token" binding:"required"`

	// Enabled is the flag to enable or disable the node
	Enabled bool `json:"enabled"`
	// Schedulable is the flag to enable or disable scheduling on the node
	Schedulable bool `json:"schedulable"`
}
