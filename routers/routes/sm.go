package routes

import (
	"go-deploy/routers/api/v1/v1_sm"
)

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
		{Method: "GET", Pattern: SMsPath, HandlerFunc: v1_sm.ListSMs},
		{Method: "GET", Pattern: SmPath, HandlerFunc: v1_sm.GetSM},
		{Method: "DELETE", Pattern: SmPath, HandlerFunc: v1_sm.DeleteSM},
	}
}
