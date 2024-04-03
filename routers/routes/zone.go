package routes

import v1 "go-deploy/routers/api/v1"

const (
	ZonesPath = "/v1/zones"
	// TODO:
	//ZonePath  = "/zones/:id"
)

type ZoneRoutingGroup struct{ RoutingGroupBase }

func ZoneRoutes() *ZoneRoutingGroup {
	return &ZoneRoutingGroup{}
}

func (group *ZoneRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: ZonesPath, HandlerFunc: v1.ListZones},
	}
}
