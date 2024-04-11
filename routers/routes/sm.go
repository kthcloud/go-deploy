package routes

import v1 "go-deploy/routers/api/v1"

const (
	SMsPath = "/v1/storageManagers"
	SmPath  = "/v1/storageManagers/:storageManagerId"
)

type SmRoutingGroup struct{ RoutingGroupBase }

func SmRoutes() *SmRoutingGroup {
	return &SmRoutingGroup{}
}

func (group *SmRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: SMsPath, HandlerFunc: v1.ListSMs},
		{Method: "GET", Pattern: SmPath, HandlerFunc: v1.GetSM},
		{Method: "DELETE", Pattern: SmPath, HandlerFunc: v1.DeleteSM},
	}
}
