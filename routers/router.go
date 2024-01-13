package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/penglongli/gin-metrics/ginmetrics"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go-deploy/docs"
	"go-deploy/pkg/auth"
	"go-deploy/pkg/config"
	"go-deploy/pkg/metrics"
	"go-deploy/pkg/sys"
	"go-deploy/routers/api/v1/middleware"
	"go-deploy/routers/api/validators"
	"go-deploy/routers/routes"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

func NewRouter() *gin.Engine {
	router := gin.New()
	basePath := getUrlBasePath()

	// Global middleware
	router.Use(CorsAllowAll())
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Metrics middleware
	m := ginmetrics.GetMonitor()
	m.SetMetricPath("/internal/metrics")
	m.SetMetricPrefix(metrics.Prefix)
	m.Use(router)

	// Index routes
	router.StaticFile("static/favicon.ico", "index/static/favicon.ico")
	router.StaticFile("static/style.css", "index/static/style.css")
	router.StaticFile("static/logo.png", "index/static/logo.png")
	router.StaticFile("static/script.js", "index/static/script.js")
	router.LoadHTMLFiles("index/index.html")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"staticFolder": basePath + "/static",
			"apiUrl":       config.Config.ExternalUrl + "/v1/status",
		})
	})

	// Private routing group
	private := router.Group("/")
	private.Use(auth.New(auth.Check(), sys.GetKeyCloakConfig()))
	private.Use(middleware.SynchronizeUser)
	private.Use(middleware.UserHttpEvent())

	// Public routing group
	public := router.Group("/")

	//// Swagger routes
	docs.SwaggerInfo.BasePath = basePath + "/v1"
	public.GET("/v1/docs", func(c *gin.Context) {
		c.Redirect(302, "/v1/docs/index.html")
	})
	public.GET("/v1/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	//// Health check routes
	public.GET("/healthz", func(c *gin.Context) {
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

func HandleRoute(engine *gin.RouterGroup, method, path string, handler gin.HandlerFunc, middleware []gin.HandlerFunc) {
	allHandlers := append(middleware, handler)
	engine.Handle(method, path, allHandlers...)
}

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
			"custom_domain":          validators.CustomDomain,
			"health_check_path":      validators.HealthCheckPath,
			"team_name":              validators.TeamName,
			"team_member_list":       validators.TeamMemberList,
			"team_resource_list":     validators.TeamResourceList,
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
		log.Fatal(err)
	}
	res = u.Path

	// Remove trailing slash
	if strings.HasSuffix(res, "/") {
		res = strings.TrimSuffix(res, "/")
	}

	return res
}
