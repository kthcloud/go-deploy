package routes

import (
	"go-deploy/routers/api/v1/v1_status"
)

const (
	StatusPath = "/v1/status"
)

type StatusRoutingGroup struct{ RoutingGroupBase }

func StatusRoutes() *StatusRoutingGroup {
	return &StatusRoutingGroup{}
}

func (group *StatusRoutingGroup) PublicRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: StatusPath, HandlerFunc: v1_status.List},
	}
}
