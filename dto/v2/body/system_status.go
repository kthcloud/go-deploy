package body

import (
	"go-deploy/pkg/subsystems/host_api"
	"time"
)

type SystemStatus struct {
	Hosts []HostStatus `json:"hosts" bson:"hosts"`
}

type TimestampedSystemStatus struct {
	Status    SystemStatus `json:"systemStatus" bson:"systemStatus"`
	Timestamp time.Time    `json:"timestamp" bson:"timestamp"`
}

type HostStatus struct {
	HostBase        `json:",inline" bson:",inline"`
	host_api.Status `json:",inline" bson:",inline"`
}
