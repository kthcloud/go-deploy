package routes

import "github.com/kthcloud/go-deploy/routers/api/v2"

const (
	VmActionsPath = "/v2/vmActions"
	VmActionPath  = "/v2/vmActions/:actionId"
)

type VmActionRoutingGroup struct{ RoutingGroupBase }

func VmActionRoutes() *VmActionRoutingGroup {
	return &VmActionRoutingGroup{}
}

func (group *VmActionRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "POST", Pattern: VmActionsPath, HandlerFunc: v2.CreateVmAction},
	}
}
