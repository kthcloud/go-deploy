package routers

import (
	"github.com/gin-gonic/gin"
	"go-deploy/pkg/app"
	"go-deploy/pkg/auth"
	"go-deploy/routers/api/v1/v1_deployment"
	"go-deploy/routers/api/v1/v1_vm"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	apiv1 := r.Group("/v1")
	apiv1.Use(auth.New(auth.Check(), app.GetKeyCloakConfig()))

	apiv1Hook := r.Group("/v1/hooks")

	setupDeploymentRoutes(apiv1, apiv1Hook)
	setupVMRoutes(apiv1, apiv1Hook)

	return r
}

func setupDeploymentRoutes(base *gin.RouterGroup, hooks *gin.RouterGroup) {
	base.GET("/deployments", v1_deployment.GetMany)

	base.GET("/deployments/:deploymentId", v1_deployment.Get)
	base.GET("/deployments/:deploymentId/ciConfig", v1_deployment.GetCiConfig)
	base.GET("/deployments/:deploymentId/logs", v1_deployment.GetLogs)
	base.POST("/deployments", v1_deployment.Create)
	base.DELETE("/deployments/:deploymentId", v1_deployment.Delete)

	hooks.POST("/deployments/harbor", v1_deployment.HandleHarborHook)

}

func setupVMRoutes(base *gin.RouterGroup, _ *gin.RouterGroup) {
	base.GET("/vms", v1_vm.GetMany)

	base.GET("/vms/:vmId", v1_vm.Get)
	base.POST("/vms", v1_vm.Create)
	base.POST("/vms/:vmId/keyPair", v1_vm.CreateKeyPair)
	base.POST("/vms/:vmId/command", v1_vm.DoCommand)
	base.DELETE("/vms/:vmId", v1_vm.Delete)
}
