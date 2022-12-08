package v1

import (
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	"go-deploy/service/deployment_service"
	"go-deploy/utils"
	"log"
	"net/http"
	"strings"
)

func getTokenFromAuthHeader(context app.ClientContext) (string, error) {
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

func HandleHarborHook(c *gin.Context) {
	context := app.ClientContext{GinContext: c}

	token, err := getTokenFromAuthHeader(context)

	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if token == "" {
		context.Unauthorized()
		return
	}

	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	deployment, err := deployment_service.GetByWebhookToken(utils.HashString(token))
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if deployment == nil {
		context.NotFound()
		return
	}

	webook, err := deployment_service.GetHook(context.GinContext.Request.Body)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if webook.Type == "PUSH_ARTIFACT" {
		log.Printf("restarting deployment %s due to push\n", deployment.Name)
		err = deployment_service.Restart(deployment.Name)
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
			return
		}
	}

	context.Ok()
}
