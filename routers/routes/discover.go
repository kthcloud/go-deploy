package routes

import v1 "go-deploy/routers/api/v1"

const (
	DiscoverPath = "/v1/discover"
)

type DiscoverRoutingGroup struct{ RoutingGroupBase }

func DiscoverRoutes() *DiscoverRoutingGroup {
	return &DiscoverRoutingGroup{}
}

func (group *DiscoverRoutingGroup) PublicRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: DiscoverPath, HandlerFunc: v1.Discover},
	}
}
