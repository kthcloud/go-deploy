package v1_deployment

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/uri"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/deployment_service"
	"net/http"
)

// GetCiConfig
// @Summary Get CI config
// @Description Get CI config
// @Tags Deployment
// @Accept  json
// @Produce  json
// @Param deploymentId path string true "Deployment ID"
// @Success 200 {object} body.CiConfig
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /api/v1/deployments/{deploymentId}/ciConfig [get]
func GetCiConfig(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.CiConfigGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	config, err := deployment_service.GetCIConfig(requestURI.DeploymentID, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if config == nil {
		context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceNotReady, fmt.Sprintf("CI config is not ready"))
		return
	}

	context.JSONResponse(http.StatusOK, config)
}
