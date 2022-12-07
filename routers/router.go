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

	apiv1User := apiv1.Group("/users/:userId")
	apiv1User.Use(auth.New(auth.Check(), app.GetKeyCloakConfig()))

	apiv1Hook := apiv1.Group("/hooks/")

	{
		// path: /
		apiv1.GET("/projects", v1.GetAllProjects)

		apiv1.GET("/projects/:projectId", v1.GetProject)
		apiv1.GET("/projects/:projectId/status", v1.GetProjectStatus)
		apiv1.GET("/projects/:projectId/ciConfigs", v1.GetCIConfig)
		apiv1.GET("/projects/:projectId/logs", v1.GetProjectLogs)
		apiv1.POST("/projects", v1.CreateProject)
		apiv1.DELETE("/projects/:projectId", v1.DeleteProject)

		{
			// path: /hooks
			apiv1Hook.POST("/projects", v1.HandleProjectHook)
		}

		{
			// path: /user/:userId/
			apiv1User.GET("/projects", v1.GetProjectsByOwnerID)
		}
	}

	return r
}
