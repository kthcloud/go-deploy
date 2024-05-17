package routes

import (
	v2 "go-deploy/routers/api/v2"
)

const (
	VmsPath = "/v2/vms"
	VmPath  = "/v2/vms/:vmId"
)

type VmRoutingGroup struct{ RoutingGroupBase }

func VmRoutes() *VmRoutingGroup {
	return &VmRoutingGroup{}
}

func (group *VmRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: VmPath, HandlerFunc: v2.GetVM},
		{Method: "GET", Pattern: VmsPath, HandlerFunc: v2.ListVMs},
		{Method: "POST", Pattern: VmsPath, HandlerFunc: v2.CreateVM},
		{Method: "POST", Pattern: VmPath, HandlerFunc: v2.UpdateVM},
		{Method: "DELETE", Pattern: VmPath, HandlerFunc: v2.DeleteVM},
	}
}
