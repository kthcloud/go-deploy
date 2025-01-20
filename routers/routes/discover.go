package routes

import v2 "github.com/kthcloud/go-deploy/routers/api/v2"

const (
	DiscoverPath = "/v2/discover"
)

type DiscoverRoutingGroup struct{ RoutingGroupBase }

func DiscoverRoutes() *DiscoverRoutingGroup {
	return &DiscoverRoutingGroup{}
}

func (group *DiscoverRoutingGroup) PublicRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: DiscoverPath, HandlerFunc: v2.Discover},
	}
}
