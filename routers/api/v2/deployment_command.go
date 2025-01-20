package v2

import (
	"github.com/gin-gonic/gin"
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/dto/v2/uri"
	"github.com/kthcloud/go-deploy/pkg/sys"
	"github.com/kthcloud/go-deploy/service"
	"github.com/kthcloud/go-deploy/service/v2/deployments/opts"
)

// DoDeploymentCommand
// @Summary Do command
// @Description Do command
// @Tags Deployment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param deploymentId path string true "Deployment ID"
// @Param body body body.DeploymentCommand true "Command body"
// @Success 204 "No Content"
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/deployments/{deploymentId}/command [post]
func DoDeploymentCommand(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.DeploymentCommand
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var requestBody body.DeploymentCommand
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	deployV2 := service.V2(auth)

	deployment, err := deployV2.Deployments().Get(requestURI.DeploymentID, opts.GetOpts{Shared: true})
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	if deployment == nil {
		context.NotFound("Deployment not found")
		return
	}

	if !deployment.Ready() {
		context.UserError("Resource is not ready")
		return
	}

	deployV2.Deployments().DoCommand(requestURI.DeploymentID, requestBody.Command)

	context.OkNoContent()
}
