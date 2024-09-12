package routers

import (
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/penglongli/gin-metrics/ginmetrics"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	docsV2 "github.com/kthcloud/go-deploy/docs/api/v2"
	"github.com/kthcloud/go-deploy/models/mode"
	"github.com/kthcloud/go-deploy/pkg/auth"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/metrics"
	"github.com/kthcloud/go-deploy/pkg/sys"
	"github.com/kthcloud/go-deploy/routers/api/v2/middleware"
	"github.com/kthcloud/go-deploy/routers/api/validators"
	"github.com/kthcloud/go-deploy/routers/routes"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
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

	// Private routing group
	private := router.Group("/")
	private.Use(auth.SetupKeycloakChain(auth.Check(), sys.GetKeyCloakConfig()))
	private.Use(middleware.SetupAuthUser)

	// Public routing group
	public := router.Group("/")

	//// Swagger routes
	// v2
	docsV2.SwaggerInfoV2.BasePath = basePath
	public.GET("/v2/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, ginSwagger.InstanceName("V2")))

	//// Health check routes
	public.Any("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// TODO: Answer /livez or/and /readyz with 200 if the server is up

	// Hook routing group
	hook := router.Group("/")

	groups := routes.RoutingGroups()
	for _, group := range groups {
		for _, route := range group.PublicRoutes() {
			HandleRoute(public, route.Method, route.Pattern, route.HandlerFunc, route.Middleware)
		}

		for _, route := range group.PrivateRoutes() {
			HandleRoute(private, route.Method, route.Pattern, route.HandlerFunc, route.Middleware)
		}

		for _, route := range group.HookRoutes() {
			HandleRoute(hook, route.Method, route.Pattern, route.HandlerFunc, route.Middleware)
		}
	}

	registerCustomValidators()

	return router
}

// HandleRoute registers a route with the given method, path, handler and middleware
func HandleRoute(engine *gin.RouterGroup, method, path string, handler gin.HandlerFunc, middleware []gin.HandlerFunc) {
	allHandlers := append(middleware, handler)
	engine.Handle(method, path, allHandlers...)
}

// registerCustomValidators registers custom validators for the gin binding
func registerCustomValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

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
	if strings.HasSuffix(res, "/") {
		res = strings.TrimSuffix(res, "/")
	}

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
