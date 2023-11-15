package routes

import (
	"go-deploy/routers/api/v1/v1_discover"
)

const (
	DiscoverPath = "/v1/discover"
)

type DiscoverRoutingGroup struct{ RoutingGroupBase }

func DiscoverRoutes() *DiscoverRoutingGroup {
	return &DiscoverRoutingGroup{}
}

func (group *DiscoverRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: DiscoverPath, HandlerFunc: v1_discover.Discover},
	}
}
