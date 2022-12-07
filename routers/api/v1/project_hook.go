package v1

import (
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	"go-deploy/service/project_service"
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

func HandleProjectHook(c *gin.Context) {
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

	project, err := project_service.GetByWebhookToken(utils.HashString(token))
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if project == nil {
		context.NotFound()
		return
	}

	webook, err := project_service.GetHook(context.GinContext.Request.Body)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if webook.Type == "PUSH_ARTIFACT" {
		log.Printf("restarting project %s due to push\n", project.Name)
		err = project_service.Restart(project.Name)
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
			return
		}
	}

	context.Ok()
}
