package routes

import (
	"go-deploy/routers/api/v1/v1_deployment"
)

const (
	StorageManagersPath = "/v1/storageManagers"
	StorageManagerPath  = "/v1/storageManagers/:storageManagerId"
)

type StorageManagerRoutingGroup struct{ RoutingGroupBase }

func StorageManagerRoutes() *StorageManagerRoutingGroup {
	return &StorageManagerRoutingGroup{}
}

func (group *StorageManagerRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: StorageManagersPath, HandlerFunc: v1_deployment.ListStorageManagers},
		{Method: "GET", Pattern: StorageManagerPath, HandlerFunc: v1_deployment.GetStorageManager},
		{Method: "DELETE", Pattern: StorageManagerPath, HandlerFunc: v1_deployment.DeleteStorageManager},
	}
}
