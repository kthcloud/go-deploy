package routes

import v2 "github.com/kthcloud/go-deploy/routers/api/v2"

const (
	HostsPath = "/v2/hosts"
	HostPath  = "/v2/hosts/:hostId"
)

type HostRoutingGroup struct{ RoutingGroupBase }

func HostRoutes() *HostRoutingGroup {
	return &HostRoutingGroup{}
}

func (group *HostRoutingGroup) PublicRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: HostsPath, HandlerFunc: v2.ListHosts},
	}
}
