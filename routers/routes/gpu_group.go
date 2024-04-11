package routes

import "go-deploy/routers/api/v2"

const (
	GpuGroupsPath = "/v2/gpuGroups"
	GpuGroupPath  = "/v2/gpuGroups/:gpuGroupId"
)

type GpuGroupRoutingGroup struct{ RoutingGroupBase }

func GpuGroupRoutes() *GpuGroupRoutingGroup {
	return &GpuGroupRoutingGroup{}
}

func (group *GpuGroupRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: GpuGroupPath, HandlerFunc: v2.GetGpuGroup},
		{Method: "GET", Pattern: GpuGroupsPath, HandlerFunc: v2.ListGpuGroups},
	}
}
