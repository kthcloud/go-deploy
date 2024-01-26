package v1_deployment

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/v1/body"
	"go-deploy/models/dto/v1/uri"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	"go-deploy/service/v1/deployments/opts"
	"net/http"
)

// DoCommand
// @Summary Do command
// @Description Do command
// @Tags Deployment
// @Accept  json
// @Produce  json
// @Param deploymentId path string true "Deployment ID"
// @Param body body body.DeploymentCommand true "Command body"
// @Success 200 {empty} empty
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /deployments/{deploymentId}/command [post]
func DoCommand(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.DeploymentCommand
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	var requestBody body.DeploymentCommand
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)

	deployment, err := deployV1.Deployments().Get(requestURI.DeploymentID, opts.GetOpts{Shared: true})
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if deployment == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("Resource with id %s not found", requestURI.DeploymentID))
		return
	}

	if !deployment.Ready() {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceNotReady, fmt.Sprintf("Resource %s is not ready", requestURI.DeploymentID))
		return
	}

	deployV1.Deployments().DoCommand(requestURI.DeploymentID, requestBody.Command)

	context.OkNoContent()
}
