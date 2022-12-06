package routers

import (
	"deploy-api-go/pkg/app"
	"deploy-api-go/pkg/auth"
	"deploy-api-go/routers/api/v1"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiv1 := r.Group("/api/v1")
	apiv1User := apiv1.Group("/users/:userId")

	apiv1.Use(auth.New(auth.Check(), app.GetKeyCloakConfig()))
	apiv1User.Use(auth.New(auth.Check(), app.GetKeyCloakConfig()))

	{
		// Path: /
		apiv1.GET("/projects", v1.GetAllProjects)

		// Path: /user/:userId/
		{
			apiv1User.GET("/projects", v1.GetProjectsByOwnerID)
			apiv1User.GET("/projects/:projectId", v1.GetProject)
			apiv1User.POST("/projects", v1.CreateProject)
			apiv1User.DELETE("/projects/:projectId", v1.DeleteProject)

			apiv1User.GET("/status/:projectId", v1.GetProjectStatus)
		}
	}

	return r
}
