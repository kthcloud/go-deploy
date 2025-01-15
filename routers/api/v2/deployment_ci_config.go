package v2

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kthcloud/go-deploy/dto/v2/uri"
	"github.com/kthcloud/go-deploy/pkg/sys"
	"github.com/kthcloud/go-deploy/service"
	dErrors "github.com/kthcloud/go-deploy/service/errors"
)

// GetCiConfig
// @Summary Get CI config
// @Description Get CI config
// @Tags Deployment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param deploymentId path string true "Deployment ID"
// @Success 200 {object} body.CiConfig
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/deployments/{deploymentId}/ciConfig [get]
func GetCiConfig(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.CiConfigGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	config, err := service.V2(auth).Deployments().GetCiConfig(requestURI.DeploymentID)
	if err != nil {
		if errors.Is(err, dErrors.DeploymentNotFoundErr) {
			context.NotFound("Deployment not found")
			return
		}

		if errors.Is(err, dErrors.DeploymentHasNotCiConfigErr) {
			context.UserError("Deployment has not CI config (not a custom deployment)")
			return
		}

		context.ServerError(err, InternalError)
		return
	}

	if config == nil {
		context.UserError("CI config is not ready")
		return
	}

	context.JSONResponse(http.StatusOK, config)
}
