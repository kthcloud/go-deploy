package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go-deploy/pkg/metrics"
)

const (
	MetricsPath = "/v1/metrics"
)

type MetricsRoutingGroup struct {
	RoutingGroupBase

	handler gin.HandlerFunc
}

func MetricsRoutes() *MetricsRoutingGroup {
	promHttpHandler := promhttp.HandlerFor(metrics.Metrics.Registry, promhttp.HandlerOpts{})
	ginHandler := func(c *gin.Context) {
		metrics.Sync()
		promHttpHandler.ServeHTTP(c.Writer, c.Request)
	}

	return &MetricsRoutingGroup{
		handler: ginHandler,
	}
}

func (group *MetricsRoutingGroup) PublicRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: MetricsPath, HandlerFunc: group.handler},
	}
}
