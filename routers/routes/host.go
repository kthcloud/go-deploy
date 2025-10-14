package routes

import v2 "github.com/kthcloud/go-deploy/routers/api/v2"

const (
	HostsPath        = "/v2/hosts"
	HostsPathVerbose = "/v2/hosts/verbose"
	HostPath         = "/v2/hosts/:hostId"
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

func (group *HostRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: HostsPathVerbose, HandlerFunc: v2.VerboseListHosts},
	}
}
