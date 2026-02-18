package routes

import "github.com/gin-gonic/gin"

type Route struct {
	Method      string
	Pattern     string
	HandlerFunc gin.HandlerFunc
	Middleware  []gin.HandlerFunc
}

// RoutingGroup is a group of routes
// Each RoutingGroup should be a separate file, that implements the RoutingGroup interface
// This ensures that each routing group is self-contained and can itself determine
// what routes should be public, private, or hook routes
type RoutingGroup interface {
	PublicRoutes() []Route
	PrivateRoutes() []Route
	HookRoutes() []Route
}

// RoutingGroups returns a list of all routing groups that should be registered in the router
func RoutingGroups() []RoutingGroup {
	return []RoutingGroup{
		DiscoverRoutes(),
		DeploymentRoutes(),
		GpuClaimRoutes(),
		GpuGroupRoutes(),
		GpuLeaseRoutes(),
		HostRoutes(),
		JobRoutes(),
		MetricsRoutes(),
		NotificationRoutes(),
		RegisterRoutes(),
		ResourceMigrationRoutes(),
		SmRoutes(),
		SnapshotRoutes(),
		SystemRoutes(),
		TeamRoutes(),
		UserRoutes(),
		VmActionRoutes(),
		VmRoutes(),
		ZoneRoutes(),
	}
}

// RoutingGroupBase is a base struct that implements the RoutingGroup interface
// This is useful when creating a new RoutingGroup, as it allows you to only implement the methods you need
// For example, if you only need to implement PrivateRoutes, you can do:
//
//	type MyRoutingGroup struct {
//	    RoutingGroupBase
//	}
//
//	func (group *MyRoutingGroup) PrivateRoutes() []Route {
//	    return []Route{
//	        {Method: "GET", Pattern: "/my-route", HandlerFunc: myHandler},
//	    }
//	}
type RoutingGroupBase struct{}

func (group *RoutingGroupBase) PublicRoutes() []Route  { return nil }
func (group *RoutingGroupBase) PrivateRoutes() []Route { return nil }
func (group *RoutingGroupBase) HookRoutes() []Route    { return nil }
