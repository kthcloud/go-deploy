package v1

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/dto/v1/body"
	"go-deploy/models/model"
	"go-deploy/pkg/sys"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/v1/deployments/opts"
	"strings"
	"time"
)

// HandleHarborHook
// @Summary Handle Harbor hook
// @Description Handle Harbor hook
// @Tags Deployment
// @Accept  json
// @Param Authorization header string false "Basic auth token"
// @Param body body body.HarborWebhook true "Harbor webhook body"
// @Produce  json
// @Success 204 {empty} empty
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v1/hooks/harbor [post]
func HandleHarborHook(c *gin.Context) {
	context := sys.NewContext(c)

	token, err := getHarborTokenFromAuthHeader(context)

	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if token == "" {
		context.Unauthorized("Missing token")
		return
	}

	deployV1 := service.V1()

	if !deployV1.Deployments().ValidateHarborToken(token) {
		context.Unauthorized("Invalid token")
		return
	}

	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	var webhook body.HarborWebhook
	err = context.GinContext.ShouldBindJSON(&webhook)
	if err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	deployment, err := deployV1.Deployments().Get("", opts.GetOpts{HarborWebhook: &webhook})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if deployment == nil {
		context.NotFound("Deployment not found")
		return
	}

	if webhook.Type == "PUSH_ARTIFACT" {
		newLog := model.Log{
			Source: model.LogSourceDeployment,
			Prefix: "[deployment]",
			// Since this is sent as a string, and not a JSON object, we need to prepend the createdAt
			Line:      fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), "Received push event from Registry"),
			CreatedAt: time.Now(),
		}

		deployV1.Deployments().AddLogs(deployment.ID, newLog)

		err = deployV1.Deployments().Restart(deployment.ID)
		if err != nil {
			var failedToStartActivityErr *sErrors.FailedToStartActivityError
			if errors.As(err, &failedToStartActivityErr) {
				context.Locked(failedToStartActivityErr.Error())
				return
			}

			if errors.Is(err, sErrors.DeploymentNotFoundErr) {
				context.NotFound("Deployment not found")
				return
			}

			context.ServerError(err, InternalError)
			return
		}
	}

	context.OkNoContent()
}

// getHarborTokenFromAuthHeader returns the Harbor token from the Authorization header.
func getHarborTokenFromAuthHeader(context sys.ClientContext) (string, error) {
	const authHeaderName = "Authorization"

	authHeader := context.GinContext.GetHeader(authHeaderName)
	if len(authHeader) == 0 {
		return "", nil
	}

	headerSplit := strings.Split(authHeader, " ")
	if len(headerSplit) != 2 {
		return "", nil
	}

	if headerSplit[0] != "Basic" {
		return "", nil
	}

	decodedHeader, err := base64.StdEncoding.DecodeString(headerSplit[1])
	if err != nil {
		return "", err
	}

	basicAuthSplit := strings.Split(string(decodedHeader), ":")
	if len(basicAuthSplit) != 2 {
		return "", nil
	}

	return basicAuthSplit[1], nil
}
