package routes

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/pkg/config"
	"go-deploy/pkg/metrics"
	"net/http"
	"net/http/httputil"
)

const (
	MetricsPath = "/v1/metrics"
)

type MetricsRoutingGroup struct {
	RoutingGroupBase

	handler gin.HandlerFunc
}

func MetricsRoutes() *MetricsRoutingGroup {
	target := fmt.Sprintf("localhost:%d", config.Config.Port)

	ginHandler := func(c *gin.Context) {
		metrics.Sync()

		director := func(req *http.Request) {
			req.URL.Path = "/internal/metrics"
			req.URL.Scheme = "http"
			req.URL.Host = target
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(c.Writer, c.Request)
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
