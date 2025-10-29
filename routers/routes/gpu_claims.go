package routes

import v2 "github.com/kthcloud/go-deploy/routers/api/v2"

const (
	GpuClaimsPath = "/v2/gpuClaims"
	GpuClaimPath  = "/v2/gpuClaims/:gpuClaimId"
)

type GpuClaimRoutingGroup struct{ RoutingGroupBase }

func GpuClaimRoutes() *GpuClaimRoutingGroup {
	return &GpuClaimRoutingGroup{}
}

func (group *GpuClaimRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: GpuClaimPath, HandlerFunc: v2.GetGpuClaim},
		{Method: "GET", Pattern: GpuClaimsPath, HandlerFunc: v2.ListGpuClaims},
		{Method: "POST", Pattern: GpuClaimsPath, HandlerFunc: v2.CreateGpuClaim},
		{Method: "DELETE", Pattern: GpuClaimPath, HandlerFunc: v2.DeleteDeployment},
	}
}
