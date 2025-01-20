package routes

import v2 "github.com/kthcloud/go-deploy/routers/api/v2"

const (
	ResourceMigrationsPath = "/v2/resourceMigrations"
	ResourceMigrationPath  = "/v2/resourceMigrations/:resourceMigrationId"
)

type ResourceMigrationRoutingGroup struct{ RoutingGroupBase }

func ResourceMigrationRoutes() *ResourceMigrationRoutingGroup {
	return &ResourceMigrationRoutingGroup{}
}

func (group *ResourceMigrationRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: ResourceMigrationsPath, HandlerFunc: v2.ListResourceMigrations},
		{Method: "GET", Pattern: ResourceMigrationPath, HandlerFunc: v2.GetResourceMigration},
		{Method: "POST", Pattern: ResourceMigrationsPath, HandlerFunc: v2.CreateResourceMigration},
		{Method: "POST", Pattern: ResourceMigrationPath, HandlerFunc: v2.UpdateResourceMigration},
		{Method: "DELETE", Pattern: ResourceMigrationPath, HandlerFunc: v2.DeleteResourceMigration},
	}
}
