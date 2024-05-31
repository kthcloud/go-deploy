package routes

import v2 "go-deploy/routers/api/v2"

const (
	StatusPath = "/v2/status"
)

type StatusRoutingGroup struct{ RoutingGroupBase }

func StatusRoutes() *StatusRoutingGroup {
	return &StatusRoutingGroup{}
}

func (group *StatusRoutingGroup) PublicRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: StatusPath, HandlerFunc: v2.ListWorkerStatus},
	}
}
