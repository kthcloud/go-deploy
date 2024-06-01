package routes

import v2 "go-deploy/routers/api/v2"

const (
	SystemCapacitiesPath = "/v2/systemCapacities"
	SystemStatsPath      = "/v2/systemStats"
	SystemStatusPath     = "/v2/systemStatus"
	WorkerStatusPath     = "/v2/workerStatus"
)

type SystemRoutingGroup struct{ RoutingGroupBase }

func SystemRoutes() *SystemRoutingGroup {
	return &SystemRoutingGroup{}
}

func (group *SystemRoutingGroup) PublicRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: SystemCapacitiesPath, HandlerFunc: v2.ListSystemCapacities},
		{Method: "GET", Pattern: SystemStatsPath, HandlerFunc: v2.ListSystemStats},
		{Method: "GET", Pattern: SystemStatusPath, HandlerFunc: v2.ListSystemStatus},
		{Method: "GET", Pattern: WorkerStatusPath, HandlerFunc: v2.ListWorkerStatus},
	}
}
