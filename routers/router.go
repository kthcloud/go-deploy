package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go-deploy/docs"
	"go-deploy/pkg/auth"
	"go-deploy/pkg/sys"
	"go-deploy/routers/api/v1/middleware"
	"go-deploy/routers/api/validators"
	"go-deploy/routers/routes"
	"reflect"
	"strings"
)

func NewRouter() *gin.Engine {
	router := gin.New()

	// global middleware
	router.Use(CorsAllowAll())
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	docs.SwaggerInfo.BasePath = "/v1"

	// private routing group
	private := router.Group("/")
	private.Use(auth.New(auth.Check(), sys.GetKeyCloakConfig()))
	private.Use(middleware.SynchronizeUser)
	private.Use(middleware.UserHttpEvent())

	// public routing group
	public := router.Group("/")
	public.GET("/v1/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// hook routing group
	hook := router.Group("/hooks")

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
