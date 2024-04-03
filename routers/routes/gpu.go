package routes

import (
	"github.com/gin-gonic/gin"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/routers/api/v1/middleware"
)

const (
	GpusPath = "/v1/gpus"
	// TODO:
	//GpuPath  = "/v1/gpus/:id"
)

type GpuRoutingGroup struct{ RoutingGroupBase }

func GpuRoutes() *GpuRoutingGroup {
	return &GpuRoutingGroup{}
}

func (group *GpuRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: GpusPath, HandlerFunc: v1.ListGPUs, Middleware: []gin.HandlerFunc{middleware.AccessGpuRoutes()}},
	}
}
