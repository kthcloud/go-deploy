package routes

import v2 "go-deploy/routers/api/v2"

const (
	SMsPath = "/v2/storageManagers"
	SmPath  = "/v2/storageManagers/:storageManagerId"
)

type SmRoutingGroup struct{ RoutingGroupBase }

func SmRoutes() *SmRoutingGroup {
	return &SmRoutingGroup{}
}

func (group *SmRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: SMsPath, HandlerFunc: v2.ListSMs},
		{Method: "GET", Pattern: SmPath, HandlerFunc: v2.GetSM},
		{Method: "DELETE", Pattern: SmPath, HandlerFunc: v2.DeleteSM},
	}
}
