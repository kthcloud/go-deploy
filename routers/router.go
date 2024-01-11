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
	"net/http"
	"reflect"
	"strings"
)

func NewRouter() *gin.Engine {
	router := gin.New()

	// global middleware
	router.Use(CorsAllowAll())
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.StaticFile("static/favicon.ico", "routers/static/favicon.ico")
	router.StaticFile("static/style.css", "routers/static/style.css")
	router.StaticFile("static/logo.png", "routers/static/logo.png")
	router.LoadHTMLFiles("routers/index.html")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	// metrics middleware
	m := ginmetrics.GetMonitor()
	m.SetMetricPath("/internal/metrics")
	m.SetMetricPrefix(metrics.Prefix)
	m.Use(router)

	// private routing group
	private := router.Group("/")
	private.Use(auth.New(auth.Check(), sys.GetKeyCloakConfig()))
	private.Use(middleware.SynchronizeUser)
	private.Use(middleware.UserHttpEvent())

	// public routing group
	public := router.Group("/")

	// If the public URL contains a path, it must be prepended to the swagger base path
	swaggerBase := ""
	withoutHTTPs, _ := strings.CutPrefix(config.Config.ExternalUrl, "https://")
	split := strings.SplitN(withoutHTTPs, "/", 2)
	if len(split) > 1 {
		swaggerBase = "/" + split[1]
	}
	docs.SwaggerInfo.BasePath = swaggerBase + "/v1"

	public.GET("/v1/docs", func(c *gin.Context) {
		c.Redirect(302, "/v1/docs/index.html")
	})
	public.GET("/v1/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// hook routing group
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
