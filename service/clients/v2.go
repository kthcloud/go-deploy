package clients

import (
	"go-deploy/service/core"
	apiV2 "go-deploy/service/v2/api"
)

type V2 interface {
	Auth() *core.AuthInfo
	HasAuth() bool

	Deployments() apiV2.Deployments
	Discovery() apiV2.Discovery
	Events() apiV2.Events
	Jobs() apiV2.Jobs
	Notifications() apiV2.Notifications
	SMs() apiV2.SMs
	Teams() apiV2.Teams
	Users() apiV2.Users
	ResourceMigrations() apiV2.ResourceMigrations
	System() apiV2.System
	VMs() apiV2.VMs
}
