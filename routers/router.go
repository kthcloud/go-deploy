package routers

import (
	"github.com/gin-gonic/gin"
	"go-deploy/pkg/app"
	"go-deploy/pkg/auth"
	"go-deploy/routers/api/v1"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	apiv1 := r.Group("/v1")
	apiv1.Use(auth.New(auth.Check(), app.GetKeyCloakConfig()))

	apiv1Hook := r.Group("/v1/hooks")

	{
		// path: /
		apiv1.GET("/deployments", v1.GetDeployments)

		apiv1.GET("/deployments/:deploymentId", v1.GetDeployment)
		apiv1.GET("/deployments/:deploymentId/status", v1.GetDeploymentStatus)
		apiv1.GET("/deployments/:deploymentId/ciConfig", v1.GetDeploymentCiConfig)
		apiv1.GET("/deployments/:deploymentId/logs", v1.GetDeploymentLogs)
		apiv1.POST("/deployments", v1.CreateDeployment)
		apiv1.DELETE("/deployments/:deploymentId", v1.DeleteDeployment)

		{
			// path: /hooks
			apiv1Hook.POST("/deployments/harbor", v1.HandleHarborHook)
		}
	}

	return r
}
