package routes

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/metrics"
	"net/http/httputil"
)

const (
	MetricsPath = "/v2/metrics"
)

type MetricsRoutingGroup struct {
	RoutingGroupBase

	handler gin.HandlerFunc
}

func MetricsRoutes() *MetricsRoutingGroup {
	target := fmt.Sprintf("localhost:%d", config.Config.Port)

	ginHandler := func(c *gin.Context) {
		metrics.Sync()

		rewrite := func(pr *httputil.ProxyRequest) {
			pr.Out.URL.Path = "/internal/metrics"
			pr.Out.URL.Scheme = "http"
			pr.Out.URL.Host = target
		}

		// To prevent duplicate headers, we need to clear CORS headers
		c.Writer.Header().Del("Access-Control-Allow-Origin")
		c.Writer.Header().Del("Access-Control-Allow-Methods")
		c.Writer.Header().Del("Access-Control-Allow-Credentials")

		proxy := &httputil.ReverseProxy{Rewrite: rewrite}
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
