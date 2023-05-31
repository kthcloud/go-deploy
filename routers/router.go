package routers

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"go-deploy/models/dto/body"
	"go-deploy/pkg/app"
	"go-deploy/pkg/auth"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/routers/api/v1/v1_deployment"
	"go-deploy/routers/api/v1/v1_job"
	"go-deploy/routers/api/v1/v1_user"
	"go-deploy/routers/api/v1/v1_vm"
	"golang.org/x/crypto/ssh"
	"reflect"
	"regexp"
	"strings"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go-deploy/docs"
)

func NewRouter() *gin.Engine {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AddAllowHeaders("authorization")

	router := gin.New()
	router.Use(cors.New(corsConfig))
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	docs.SwaggerInfo.BasePath = "/v1"
	privateApiv1 := router.Group("/v1")
	privateApiv1.Use(auth.New(auth.Check(), app.GetKeyCloakConfig()))
	publicApiv1 := router.Group("/v1")
	apiv1Hook := router.Group("/v1/hooks")

	router.GET("/v1/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	setupAuthCheckRoutes(privateApiv1)
	setupDeploymentRoutes(privateApiv1, publicApiv1, apiv1Hook)
	setupVmRoutes(privateApiv1, apiv1Hook)
	setupGpuRoutes(privateApiv1, apiv1Hook)
	setupJobRoutes(privateApiv1, apiv1Hook)
	setupUserRoutes(privateApiv1, apiv1Hook)

	registerCustomValidators()

	return router
}

func setupAuthCheckRoutes(private *gin.RouterGroup) {
	private.GET("/authCheck", v1_user.AuthCheck)
}

func setupDeploymentRoutes(private *gin.RouterGroup, public *gin.RouterGroup, hooks *gin.RouterGroup) {
	private.GET("/deployments", v1_deployment.GetList)

	private.GET("/deployments/:deploymentId", v1_deployment.Get)
	private.GET("/deployments/:deploymentId/ciConfig", v1_deployment.GetCiConfig)
	private.POST("/deployments", v1_deployment.Create)
	private.POST("/deployments/:deploymentId", v1_deployment.Update)
	private.DELETE("/deployments/:deploymentId", v1_deployment.Delete)

	private.POST("/deployments/:deploymentId/command", v1_deployment.DoCommand)

	public.GET("/deployments/:deploymentId/logs", v1_deployment.GetLogs)

	hooks.POST("/deployments/harbor", v1_deployment.HandleHarborHook)
	hooks.POST("/deployments/github", v1_deployment.HandleGitHubHook)
}

func setupVmRoutes(private *gin.RouterGroup, _ *gin.RouterGroup) {
	private.GET("/vms", v1_vm.GetList)

	private.GET("/vms/:vmId", v1_vm.Get)
	private.POST("/vms", v1_vm.Create)
	private.POST("/vms/:vmId", v1_vm.Update)
	private.DELETE("/vms/:vmId", v1_vm.Delete)

	private.POST("/vms/:vmId/command", v1_vm.DoCommand)

	private.POST("/vms/:vmId/attachGpu", v1_vm.AttachGPU)
	private.POST("/vms/:vmId/attachGpu/:gpuId", v1_vm.AttachGPU)
	private.POST("/vms/:vmId/detachGpu", v1_vm.DetachGPU)
}

func setupGpuRoutes(private *gin.RouterGroup, _ *gin.RouterGroup) {
	private.GET("/gpus", v1_vm.GetGpuList)
}

func setupJobRoutes(private *gin.RouterGroup, _ *gin.RouterGroup) {
	private.GET("/job/:jobId", v1_job.Get)
}

func setupUserRoutes(private *gin.RouterGroup, _ *gin.RouterGroup) {
	private.GET("/users/:userId", v1_user.Get)
	private.GET("/users", v1_user.GetList)
	private.POST("/users/:userId", v1_user.Update)
	private.POST("/users", v1_user.Update)
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

		err := v.RegisterValidation("rfc1035", func(fl validator.FieldLevel) bool {
			name, ok := fl.Field().Interface().(string)
			if !ok {
				return false
			}

			rfc1035 := regexp.MustCompile(`^[a-zA-Z]([a-zA-Z0-9-]*[a-zA-Z0-9])?([a-zA-Z]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$`)
			return rfc1035.MatchString(name)
		})
		if err != nil {
			panic(err)
		}

		err = v.RegisterValidation("ssh_public_key", func(fl validator.FieldLevel) bool {
			publicKey, ok := fl.Field().Interface().(string)
			if !ok {
				return false
			}

			_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
			if err != nil {
				return false
			}
			return true
		})
		if err != nil {
			panic(err)
		}

		err = v.RegisterValidation("env_name", func(fl validator.FieldLevel) bool {
			name, ok := fl.Field().Interface().(string)
			if !ok {
				return false
			}

			regex := regexp.MustCompile(`^[a-zA-Z]([a-zA-Z0-9-_]*[a-zA-Z0-9])?([a-zA-Z]([a-zA-Z0-9-_]*[a-zA-Z0-9])?)*$`)
			match := regex.MatchString(name)
			return match
		})
		if err != nil {
			panic(err)
		}

		err = v.RegisterValidation("env_list", func(fl validator.FieldLevel) bool {
			envList, ok := fl.Field().Interface().([]body.Env)
			if !ok {
				return false
			}

			names := make(map[string]bool)
			for _, env := range envList {
				if _, ok := names[env.Name]; ok {
					return false
				}
				names[env.Name] = true
			}
			return true
		})
		if err != nil {
			panic(err)
		}

		err = v.RegisterValidation("port_list", func(fl validator.FieldLevel) bool {
			portList, ok := fl.Field().Interface().([]body.Port)
			if !ok {
				return false
			}

			names := make(map[string]bool)
			ports := make(map[int]bool)
			for _, port := range portList {
				if _, ok := names[port.Name]; ok {
					return false
				}
				names[port.Name] = true
				if _, ok := ports[port.Port]; ok {
					return false
				}
				ports[port.Port] = true
			}
			return true
		})
		if err != nil {
			panic(err)
		}

		err = v.RegisterValidation("extra_domain_list", func(fl validator.FieldLevel) bool {
			domainList, ok := fl.Field().Interface().([]string)
			if !ok {
				return false
			}

			rfc1035 := regexp.MustCompile(`^[a-zA-Z]([a-zA-Z0-9-]*[a-zA-Z0-9])?([a-zA-Z]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$`)

			names := make(map[string]bool)
			for _, domain := range domainList {
				if _, ok := names[domain]; ok {
					return false
				}
				names[domain] = true

				splitByDots := strings.Split(domain, ".")
				for _, split := range splitByDots {
					if len(split) > 63 {
						return false
					}

					if !rfc1035.MatchString(split) {
						return false
					}
				}

				valid := v1.IsValidDomain(domain)
				if !valid {
					return false
				}
			}

			return true
		})
		if err != nil {
			panic(err)
		}
	}
}
