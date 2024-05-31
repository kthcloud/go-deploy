package body

import (
	"go-deploy/pkg/subsystems/host_api"
	"time"
)

type SystemStatus struct {
	Hosts []HostStatus `json:"hosts" bson:"hosts"`
}

type TimestampedSystemStatus struct {
	Status    SystemStatus `json:"status" bson:"status"`
	Timestamp time.Time    `json:"timestamp" bson:"timestamp"`
}

type HostStatus struct {
	HostBase        `json:",inline" bson:",inline" tstype:",extends"`
	host_api.Status `json:",inline" bson:",inline" tstype:",extends"`
}
