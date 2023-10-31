package routers

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go-deploy/docs"
	"go-deploy/pkg/auth"
	"go-deploy/pkg/metrics"
	"go-deploy/pkg/sys"
	"go-deploy/routers/api/v1/middleware"
	"go-deploy/routers/api/v1/v1_deployment"
	"go-deploy/routers/api/v1/v1_github"
	"go-deploy/routers/api/v1/v1_job"
	"go-deploy/routers/api/v1/v1_notification"
	"go-deploy/routers/api/v1/v1_user"
	"go-deploy/routers/api/v1/v1_vm"
	"go-deploy/routers/api/v1/v1_zone"
	"go-deploy/routers/api/validators"
	"reflect"
	"strings"
)

func sseHeaderMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Next()
	}
}

func NewRouter() *gin.Engine {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowCredentials = true
	corsConfig.AddAllowHeaders("authorization")

	router := gin.New()
	router.Use(cors.New(corsConfig))
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	docs.SwaggerInfo.BasePath = "/v1"
	privateApiv1 := router.Group("/v1")
	privateApiv1.Use(auth.New(auth.Check(), sys.GetKeyCloakConfig()))
	privateApiv1.Use(middleware.SynchronizeUser)
	privateApiv1.Use(middleware.UserHttpEvent())

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
	setupNotificationRoutes(privateApiv1, apiv1Hook)
	setupGitHubRoutes(privateApiv1, apiv1Hook)
	setupMetricsRoutes(publicApiv1, apiv1Hook)

	registerCustomValidators()

	return router
}

func setupAuthCheckRoutes(private *gin.RouterGroup) {
	private.GET("/authCheck", v1_user.AuthCheck)
}

func setupDeploymentRoutes(private *gin.RouterGroup, public *gin.RouterGroup, hooks *gin.RouterGroup) {
	private.GET("/deployments", middleware.CreateStorageManager(), v1_deployment.GetList)

	private.GET("/deployments/:deploymentId", middleware.CreateStorageManager(), v1_deployment.Get)
	private.GET("/deployments/:deploymentId/ciConfig", v1_deployment.GetCiConfig)
	private.POST("/deployments", middleware.CreateStorageManager(), v1_deployment.Create)
	private.POST("/deployments/:deploymentId", v1_deployment.Update)
	private.DELETE("/deployments/:deploymentId", v1_deployment.Delete)

	private.POST("/deployments/:deploymentId/command", v1_deployment.DoCommand)

	public.GET("/deployments/:deploymentId/logs", v1_deployment.GetLogs)
	private.GET("/deployments/:deploymentId/logs-sse", sseHeaderMiddleware(), v1_deployment.GetLogsSSE)

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
	private.GET("/jobs", v1_job.GetList)
	private.POST("/jobs/:jobId", v1_job.Update)
}

func setupUserRoutes(private *gin.RouterGroup, _ *gin.RouterGroup) {
	private.GET("/users/:userId", v1_user.Get)
	private.GET("/users", v1_user.GetList)
	private.POST("/users/:userId", v1_user.Update)
	private.POST("/users", v1_user.Update)

	private.GET("/teams/:teamId", v1_user.GetTeam)
	private.GET("/teams", v1_user.GetTeamList)
	private.POST("/teams", v1_user.CreateTeam)
	private.POST("/teams/:teamId", v1_user.UpdateTeam)
	private.DELETE("/teams/:teamId", v1_user.DeleteTeam)
}

func setupNotificationRoutes(private *gin.RouterGroup, _ *gin.RouterGroup) {
	private.GET("/notifications/:notificationId", v1_notification.Get)
	private.GET("/notifications", v1_notification.GetList)
	private.POST("/notifications", v1_notification.Update)
	private.DELETE("/notifications/:notificationId", v1_notification.Delete)
}

func setupGitHubRoutes(private *gin.RouterGroup, _ *gin.RouterGroup) {
	private.GET("/github/repositories", v1_github.ListGitHubRepositories)
}

func setupMetricsRoutes(base *gin.RouterGroup, _ *gin.RouterGroup) {
	base.GET("/metrics", func(c *gin.Context) {
		metrics.Sync()
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	})
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
