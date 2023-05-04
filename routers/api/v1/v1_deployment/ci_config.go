package v1_deployment

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/uri"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/deployment_service"
	"net/http"
)

func GetCiConfig(c *gin.Context) {
	context := app.NewContext(c)

	var requestURI uri.CiConfigGet
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	config, err := deployment_service.GetCIConfig(auth.UserID, requestURI.DeploymentID, auth.IsAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if config == nil {
		context.NotFound()
		return
	}

	context.JSONResponse(http.StatusOK, config)
}
