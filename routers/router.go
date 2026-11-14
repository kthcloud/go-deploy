package routers

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"strings"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	apiV2 "github.com/kthcloud/go-deploy/api/openapi/v2"
	"github.com/kthcloud/go-deploy/models/mode"
	"github.com/kthcloud/go-deploy/pkg/auth"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/metrics"
	"github.com/kthcloud/go-deploy/pkg/sys"
	v2 "github.com/kthcloud/go-deploy/routers/api/v2"
	"github.com/kthcloud/go-deploy/routers/api/v2/middleware"
	"github.com/kthcloud/go-deploy/routers/api/validators"
	"github.com/penglongli/gin-metrics/ginmetrics"
)

// NewRouter creates a new router
// It registers all routes and middlewares
func NewRouter() *gin.Engine {
	router := gin.New()
	basePath := getUrlBasePath()

	// Global middleware
	ginLogger := log.Get("api")
	router.Use(CorsAllowAll())
	router.Use(getGinLogger())
	router.Use(ginzap.RecoveryWithZap(ginLogger.Desugar(), true))

	// Metrics middleware
	m := ginmetrics.GetMonitor()
	m.SetMetricPath("/internal/metrics")
	m.SetMetricPrefix(metrics.Prefix)
	m.Use(router)

	// Index routes
	router.StaticFile("static/favicon.ico", "index/static/favicon.ico")
	router.StaticFile("static/style.css", "index/static/style.css")
	router.StaticFile("static/logo.svg", "index/static/logo.svg")
	router.StaticFile("static/script.js", "index/static/script.js")
	router.LoadHTMLFiles("index/index.html")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"staticFolder": basePath + "/static",
			"apiUrl":       config.Config.ExternalUrl + "/v2/workerStatus",
		})
	})

	// OpenAPI docs (v2)
	router.GET("/v2/docs/*any", apiV2.ServeDocs(basePath+"/v2/docs"))

	// Health check routes
	router.Any("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// TODO: Answer /livez or/and /readyz with 200 if the server is up

	publicRoutes := map[string]struct{}{
		"GET /v2/discover":                  {},
		"GET /v2/hosts":                     {},
		"GET /v2/hosts/verbose":             {},
		"GET /v2/metrics":                   {},
		"POST /v2/register":                 {},
		"GET /v2/systemCapacities":          {},
		"GET /v2/systemStats":               {},
		"GET /v2/systemStatus":              {},
		"GET /v2/workerStatus":              {},
		"POST /v2/hooks/deployments/harbor": {},
	}
	authChain := auth.SetupKeycloakChain(auth.Check(), sys.GetKeyCloakConfig())
	authMiddleware := func(c *gin.Context) {
		if _, isPublic := publicRoutes[c.Request.Method+" "+c.FullPath()]; isPublic {
			return
		}
		authChain(c)
		if c.IsAborted() {
			return
		}
		middleware.SetupAuthUser(c)
	}

	metricsTarget := fmt.Sprintf("localhost:%d", config.Config.Port)
	metricsHandler := func(c *gin.Context) {
		metrics.Sync()

		rewrite := func(pr *httputil.ProxyRequest) {
			pr.Out.URL.Path = "/internal/metrics"
			pr.Out.URL.Scheme = "http"
			pr.Out.URL.Host = metricsTarget
		}

		c.Writer.Header().Del("Access-Control-Allow-Origin")
		c.Writer.Header().Del("Access-Control-Allow-Methods")
		c.Writer.Header().Del("Access-Control-Allow-Credentials")

		proxy := &httputil.ReverseProxy{Rewrite: rewrite}
		proxy.ServeHTTP(c.Writer, c.Request)
	}

	apiV2.RegisterHandlersWithOptions(router, v2.NewServer(metricsHandler, middleware.CreateSM(), middleware.SseSetup()), apiV2.GinServerOptions{
		Middlewares: []apiV2.MiddlewareFunc{authMiddleware},
	})

	registerCustomValidators()

	return router
}

// registerCustomValidators registers custom validators for the gin binding
func registerCustomValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name, _, _ := strings.Cut(fld.Tag.Get("json"), ",")

			if name == "-" {
				name = strings.SplitN(fld.Tag.Get("uri"), ",", 2)[0]
			}

			if name == "-" {
				name = strings.SplitN(fld.Tag.Get("form"), ",", 2)[0]
			}

			if name == "-" {
				return ""
			}

			return name
		})

		registrations := map[string]func(fl validator.FieldLevel) bool{
			"rfc1035":                validators.Rfc1035,
			"rfc1123":                validators.Rfc1123,
			"ssh_public_key":         validators.SshPublicKey,
			"env_name":               validators.EnvName,
			"env_list":               validators.EnvList,
			"port_list_names":        validators.PortListNames,
			"port_list_numbers":      validators.PortListNumbers,
			"port_list_http_proxies": validators.PortListHttpProxies,
			"domain_name":            validators.DomainName,
			"health_check_path":      validators.HealthCheckPath,
			"team_name":              validators.TeamName,
			"team_member_list":       validators.TeamMemberList,
			"team_resource_list":     validators.TeamResourceList,
			"time_in_future":         validators.TimeInFuture,
			"volume_name":            validators.VolumeName,
			"deployment_name":        validators.DeploymentName,
			"vm_name":                validators.VmName,
			"vm_port_name":           validators.VmPortName,
		}

		for tag, fn := range registrations {
			err := v.RegisterValidation(tag, fn)
			if err != nil {
				panic(err)
			}
		}
	}
}

// getUrlBasePath returns the base path of the external URL.
// Meaning if we have an external URL of https://example.com/deploy,
// this function will return "/deploy"
func getUrlBasePath() string {
	res := ""

	// Parse as URL
	u, err := url.Parse(config.Config.ExternalUrl)
	if err != nil {
		log.Fatalln("failed to parse external URL. details:", err)
	}
	res = u.Path

	// Remove trailing slash
	res = strings.TrimSuffix(res, "/")

	return res
}

// getGinLogger returns the logger used for Gin Gonic.
// When in development mode, it will use the default gin.Logger(), since it is easier to read.
// When in production mode, it will use the logger from the log package.
func getGinLogger() gin.HandlerFunc {
	if config.Config.Mode != mode.Prod {
		return gin.Logger()
	}

	return ginzap.Ginzap(log.Get("api").Desugar(), time.RFC3339, true)
}
