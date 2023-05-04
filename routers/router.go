package routers

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"go-deploy/pkg/app"
	"go-deploy/pkg/auth"
	"go-deploy/routers/api/v1/v1_deployment"
	"go-deploy/routers/api/v1/v1_job"
	"go-deploy/routers/api/v1/v1_user"
	"go-deploy/routers/api/v1/v1_vm"
	"golang.org/x/crypto/ssh"
	"regexp"
)

func NewRouter() *gin.Engine {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AddAllowHeaders("authorization")

	router := gin.New()
	router.Use(cors.New(corsConfig))
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	apiv1 := router.Group("/v1")
	apiv1.Use(auth.New(auth.Check(), app.GetKeyCloakConfig()))

	apiv1Hook := router.Group("/v1/hooks")

	setupDeploymentRoutes(apiv1, apiv1Hook)
	setupVmRoutes(apiv1, apiv1Hook)
	setupGpuRoutes(apiv1, apiv1Hook)
	setupJobRoutes(apiv1, apiv1Hook)
	setupUserRoutes(apiv1, apiv1Hook)

	registerCustomValidators()

	return router
}

func setupDeploymentRoutes(base *gin.RouterGroup, hooks *gin.RouterGroup) {
	base.GET("/deployments", v1_deployment.GetList)

	base.GET("/deployments/:deploymentId", v1_deployment.Get)
	base.GET("/deployments/:deploymentId/ciConfig", v1_deployment.GetCiConfig)
	base.GET("/deployments/:deploymentId/logs", v1_deployment.GetLogs)
	base.POST("/deployments", v1_deployment.Create)
	//base.POST("/deployments/:deploymentId", v1_deployment.Update)
	base.DELETE("/deployments/:deploymentId", v1_deployment.Delete)

	hooks.POST("/deployments/harbor", v1_deployment.HandleHarborHook)
}

func setupVmRoutes(base *gin.RouterGroup, _ *gin.RouterGroup) {
	base.GET("/vms", v1_vm.GetList)

	base.GET("/vms/:vmId", v1_vm.Get)
	base.POST("/vms", v1_vm.Create)
	base.POST("/vms/:vmId", v1_vm.Update)
	base.DELETE("/vms/:vmId", v1_vm.Delete)

	base.POST("/vms/:vmId/command", v1_vm.DoCommand)

	base.POST("/vms/:vmId/attachGpu", v1_vm.AttachGPU)
	base.POST("/vms/:vmId/attachGpu/:gpuId", v1_vm.AttachGPU)
	base.POST("/vms/:vmId/detachGpu", v1_vm.DetachGPU)
}

func setupGpuRoutes(base *gin.RouterGroup, _ *gin.RouterGroup) {
	base.GET("/gpus", v1_vm.GetGpuList)
}

func setupJobRoutes(base *gin.RouterGroup, _ *gin.RouterGroup) {
	base.GET("/jobs/:jobId", v1_job.Get)
}

func setupUserRoutes(base *gin.RouterGroup, _ *gin.RouterGroup) {
	base.GET("/users/:userId", v1_user.Get)
	base.GET("/users", v1_user.GetList)
	base.POST("/users/:userId", v1_user.Update)
	base.POST("/users", v1_user.Update)
}

func validateRfc1035(fl validator.FieldLevel) bool {
	name, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	regex := regexp.MustCompile(`^[a-zA-Z]([a-zA-Z0-9-]*[a-zA-Z0-9])?([a-zA-Z]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$`)
	return regex.MatchString(name)
}

func validateSshPublicKey(fl validator.FieldLevel) bool {
	publicKey, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
	if err != nil {
		return false
	}
	return true
}

func registerCustomValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("rfc1035", validateRfc1035)
		if err != nil {
			panic(err)
		}

		err = v.RegisterValidation("ssh_public_key", validateSshPublicKey)
		if err != nil {
			panic(err)
		}
	}
}
