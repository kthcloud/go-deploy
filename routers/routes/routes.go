package routes

import "github.com/gin-gonic/gin"

type Route struct {
	Method      string
	Pattern     string
	HandlerFunc gin.HandlerFunc
	Middleware  []gin.HandlerFunc
}

type RoutingGroup interface {
	PublicRoutes() []Route
	PrivateRoutes() []Route
	HookRoutes() []Route
}

func RoutingGroups() []RoutingGroup {
	return []RoutingGroup{
		DeploymentRoutes(),
		GitHubRoutes(),
		GpuRoutes(),
		JobRoutes(),
		MetricsRoutes(),
		NotificationRoutes(),
		StorageManagerRoutes(),
		TeamRoutes(),
		UserRoutes(),
		VmRoutes(),
		ZoneRoutes(),
	}
}

type RoutingGroupBase struct{}

func (group *RoutingGroupBase) PublicRoutes() []Route  { return nil }
func (group *RoutingGroupBase) PrivateRoutes() []Route { return nil }
func (group *RoutingGroupBase) HookRoutes() []Route    { return nil }
