package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/validator"
	"go-deploy/service/project_service"
	"net/http"
)

func GetCIConfig(c *gin.Context) {
	context := app.NewContext(c)

	rules := validator.MapData{
		"userId":    []string{"required", "uuid_v4"},
		"projectId": []string{"required", "uuid_v4"},
	}

	validationErrors := context.ValidateParams(&rules)

	if len(validationErrors) > 0 {
		context.ResponseValidationError(validationErrors)
		return
	}

	userID := context.GinContext.Param("userId")
	projectID := context.GinContext.Param("projectId")
	token, err := context.GetKeycloakToken()
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if userID != token.Sub {
		context.Unauthorized()
		return
	}

	config, err := project_service.GetCIConfig(userID, projectID)
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
