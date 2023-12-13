package v1_deployment

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/uri"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/deployment_service"
	dErrors "go-deploy/service/errors"
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
// @Router /deployments/{deploymentId}/ciConfig [get]
func GetCiConfig(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.CiConfigGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	config, err := deployment_service.New().WithAuth(auth).GetCiConfig(requestURI.DeploymentID)
	if err != nil {
		if errors.Is(err, dErrors.DeploymentNotFoundErr) {
			context.NotFound("Deployment not found")
			return
		}

		context.ServerError(err, v1.InternalError)
		return
	}

	if config == nil {
		context.UserError("CI config is not ready")
		return
	}

	context.JSONResponse(http.StatusOK, config)
}
