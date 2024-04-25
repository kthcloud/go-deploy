package routes

import v1 "go-deploy/routers/api/v1"

const (
	ResourceMigrationsPath = "/v1/resourceMigrations"
	ResourceMigrationPath  = "/v1/resourceMigrations/:resourceMigrationId"
)

type ResourceMigrationRoutingGroup struct{ RoutingGroupBase }

func ResourceMigrationRoutes() *ResourceMigrationRoutingGroup {
	return &ResourceMigrationRoutingGroup{}
}

func (group *ResourceMigrationRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: ResourceMigrationsPath, HandlerFunc: v1.ListResourceMigrations},
		{Method: "GET", Pattern: ResourceMigrationPath, HandlerFunc: v1.GetResourceMigration},
		{Method: "POST", Pattern: ResourceMigrationsPath, HandlerFunc: v1.CreateResourceMigration},
		{Method: "POST", Pattern: ResourceMigrationPath, HandlerFunc: v1.UpdateResourceMigration},
		{Method: "DELETE", Pattern: ResourceMigrationPath, HandlerFunc: v1.DeleteResourceMigration},
	}
}
