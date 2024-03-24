package routes

import "go-deploy/routers/api/v2"

const (
	GpuLeasesPath = "/v2/gpuLeases"
	GpuLeasePath  = "/v2/gpuLeases/:gpuLeaseId"
)

type GpuLeaseRoutingGroup struct{ RoutingGroupBase }

func GpuLeaseRoutes() *GpuLeaseRoutingGroup {
	return &GpuLeaseRoutingGroup{}
}

func (group *GpuLeaseRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: GpuLeasePath, HandlerFunc: v2.GetGpuLease},
		{Method: "GET", Pattern: GpuLeasesPath, HandlerFunc: v2.ListGpuLeases},
		{Method: "POST", Pattern: GpuLeasesPath, HandlerFunc: v2.CreateGpuLease},
		{Method: "DELETE", Pattern: GpuLeasePath, HandlerFunc: v2.DeleteGpuLease},
	}
}
