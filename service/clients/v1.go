package clients

import (
	"go-deploy/service/core"
	apiV1 "go-deploy/service/v1/api"
)

type V1 interface {
	Auth() *core.AuthInfo
	HasAuth() bool

	Deployments() apiV1.Deployments
	Discovery() apiV1.Discovery
	Events() apiV1.Events
	Jobs() apiV1.Jobs
	Notifications() apiV1.Notifications
	SMs() apiV1.SMs
	Status() apiV1.Status
	Teams() apiV1.Teams
	Users() apiV1.Users
	UserData() apiV1.UserData
	VMs() apiV1.VMs
	Zones() apiV1.Zones
}
