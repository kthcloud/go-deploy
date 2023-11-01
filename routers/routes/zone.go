package routes

import "go-deploy/routers/api/v1/v1_zone"

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
		{Method: "GET", Pattern: ZonesPath, HandlerFunc: v1_zone.List},
	}
}
