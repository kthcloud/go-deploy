package v1_deployment

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/validator"
	"go-deploy/service/deployment_service"
	"net/http"
)

func GetCiConfig(c *gin.Context) {
	context := app.NewContext(c)

	rules := validator.MapData{
		"deploymentId": []string{"required", "uuid_v4"},
	}

	validationErrors := context.ValidateParams(&rules)

	if len(validationErrors) > 0 {
		context.ResponseValidationError(validationErrors)
		return
	}

	token, err := context.GetKeycloakToken()
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}
	userID := token.Sub
	deploymentID := context.GinContext.Param("deploymentId")

	config, err := deployment_service.GetCIConfig(userID, deploymentID)
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
