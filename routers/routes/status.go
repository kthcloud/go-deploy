package routes

import v1 "go-deploy/routers/api/v1"

const (
	StatusPath = "/v1/status"
)

type StatusRoutingGroup struct{ RoutingGroupBase }

func StatusRoutes() *StatusRoutingGroup {
	return &StatusRoutingGroup{}
}

func (group *StatusRoutingGroup) PublicRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: StatusPath, HandlerFunc: v1.ListWorkerStatus},
	}
}
