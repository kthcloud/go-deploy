package routers

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go-deploy/pkg/app"
	"go-deploy/pkg/auth"
	"go-deploy/routers/api/v1"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiv1 := r.Group("/api/v1")

	apiv1User := apiv1.Group("/users/:userId")
	apiv1User.Use(auth.New(auth.Check(), app.GetKeyCloakConfig()))

	apiv1Hook := apiv1.Group("/hooks/")

	{
		// path: /
		apiv1.GET("/projects", v1.GetAllProjects)

		{
			// path: /hooks
			apiv1Hook.POST("/projects", v1.HandleProjectHook)
		}

		{
			// path: /user/:userId/
			apiv1User.GET("/projects", v1.GetProjectsByOwnerID)
			apiv1User.GET("/projects/:projectId", v1.GetProject)
			apiv1User.POST("/projects", v1.CreateProject)
			apiv1User.DELETE("/projects/:projectId", v1.DeleteProject)

			apiv1User.GET("/status/:projectId", v1.GetProjectStatus)

			apiv1User.GET("/ciConfigs/:projectId/", v1.GetCIConfig)

			apiv1User.GET("/logs/:projectId", v1.GetProjectLogs)
		}
	}

	return r
}
