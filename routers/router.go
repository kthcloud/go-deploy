package routers

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"go-deploy/models/dto/body"
	"go-deploy/pkg/auth"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/routers/api/v1/v1_deployment"
	"go-deploy/routers/api/v1/v1_github"
	"go-deploy/routers/api/v1/v1_job"
	"go-deploy/routers/api/v1/v1_user"
	"go-deploy/routers/api/v1/v1_vm"
	"go-deploy/routers/api/v1/v1_zone"
	"golang.org/x/crypto/ssh"
	"reflect"
	"regexp"
	"strconv"
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
	privateApiv1.Use(auth.New(auth.Check(), sys.GetKeyCloakConfig()))
	privateApiv1.Use(v1_user.SynchronizeUser)

	publicApiv1 := router.Group("/v1")
	apiv1Hook := router.Group("/v1/hooks")

	router.GET("/v1/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	setupAuthCheckRoutes(privateApiv1)
	setupDeploymentRoutes(privateApiv1, publicApiv1, apiv1Hook)
	setupStorageManagerRoutes(privateApiv1, apiv1Hook)
	setupVmRoutes(privateApiv1, apiv1Hook)
	setupZoneRoutes(privateApiv1, apiv1Hook)
	setupGpuRoutes(privateApiv1, apiv1Hook)
	setupJobRoutes(privateApiv1, apiv1Hook)
	setupUserRoutes(privateApiv1, apiv1Hook)
	setupGitHubRoutes(privateApiv1, apiv1Hook)

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

func setupStorageManagerRoutes(private *gin.RouterGroup, _ *gin.RouterGroup) {
	private.GET("/storageManagers", v1_deployment.GetStorageManagerList)

	private.GET("/storageManagers/:storageManagerId", v1_deployment.GetStorageManager)
	private.DELETE("/storageManagers/:storageManagerId", v1_deployment.DeleteStorageManager)
}

func setupVmRoutes(private *gin.RouterGroup, _ *gin.RouterGroup) {
	private.GET("/vms", v1_vm.GetList)

	private.GET("/vms/:vmId", v1_vm.Get)
	private.POST("/vms", v1_vm.Create)
	private.POST("/vms/:vmId", v1_vm.Update)
	private.DELETE("/vms/:vmId", v1_vm.Delete)

	private.GET("/vms/:vmId/snapshots", v1_vm.GetSnapshotList)
	private.POST("/vms/:vmId/snapshots", v1_vm.CreateSnapshot)

	private.POST("/vms/:vmId/command", v1_vm.DoCommand)

	gpuRoutes := private.Group("/vms/:vmId")
	gpuRoutes.Use(v1_vm.AccessGpuRoutes)
}

func setupZoneRoutes(private *gin.RouterGroup, _ *gin.RouterGroup) {
	private.GET("/zones", v1_zone.GetList)
}

func setupGpuRoutes(private *gin.RouterGroup, _ *gin.RouterGroup) {
	gpuRoutes := private.Group("/")
	gpuRoutes.Use(v1_vm.AccessGpuRoutes)
	gpuRoutes.GET("/gpus", v1_vm.GetGpuList)
}

func setupJobRoutes(private *gin.RouterGroup, _ *gin.RouterGroup) {
	private.GET("/jobs/:jobId", v1_job.Get)
}

func setupUserRoutes(private *gin.RouterGroup, _ *gin.RouterGroup) {
	private.GET("/users/:userId", v1_user.Get)
	private.GET("/users", v1_user.GetList)
	private.POST("/users/:userId", v1_user.Update)
	private.POST("/users", v1_user.Update)
}

func setupGitHubRoutes(private *gin.RouterGroup, _ *gin.RouterGroup) {
	private.GET("/github/repositories", v1_github.ListGitHubRepositories)
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

		err = v.RegisterValidation("port_list_names", func(fl validator.FieldLevel) bool {
			portList, ok := fl.Field().Interface().([]body.Port)
			if !ok {
				return false
			}

			names := make(map[string]bool)
			for _, port := range portList {
				if _, ok := names[port.Name]; ok {
					return false
				}
				names[port.Name] = true
			}
			return true
		})
		if err != nil {
			panic(err)
		}

		err = v.RegisterValidation("port_list_numbers", func(fl validator.FieldLevel) bool {
			portList, ok := fl.Field().Interface().([]body.Port)
			if !ok {
				return false
			}

			ports := make(map[string]bool)
			for _, port := range portList {
				identifier := strconv.Itoa(port.Port) + "/" + port.Protocol
				if _, ok := ports[identifier]; ok {
					return false
				}
				ports[identifier] = true
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

				illegalSuffixes := make([]string, len(conf.Env.Deployment.Zones))
				for idx, zone := range conf.Env.Deployment.Zones {
					illegalSuffixes[idx] = zone.ParentDomain
				}

				for _, suffix := range illegalSuffixes {
					if strings.HasSuffix(domain, suffix) {
						return false
					}
				}

				if !v1.IsValidDomain(domain) {
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
