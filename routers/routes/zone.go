package routes

import v2 "github.com/kthcloud/go-deploy/routers/api/v2"

const (
	ZonesPath = "/v2/zones"
	// TODO:
	//ZonePath  = "/zones/:id"
)

type ZoneRoutingGroup struct{ RoutingGroupBase }

func ZoneRoutes() *ZoneRoutingGroup {
	return &ZoneRoutingGroup{}
}

func (group *ZoneRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: ZonesPath, HandlerFunc: v2.ListZones},
	}
}
