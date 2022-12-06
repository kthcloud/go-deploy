package v1

import (
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/validator"
	"go-deploy/service/project_service"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetProjectStatus(c *gin.Context) {
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

	userId := context.GinContext.Param("userId")
	projectId := context.GinContext.Param("projectId")
	token, err := context.GetKeycloakToken()
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
	}

	if userId != token.Sub {
		context.Unauthorized()
		return
	}

	statusCode, projectStatus, _ := project_service.GetStatusByID(userId, projectId)

	if projectStatus == nil {
		context.NotFound()
		return
	}

	if statusCode == status_codes.ProjectNotFound {
		context.JsonResponse(http.StatusNotFound, projectStatus)
		return
	}

	context.JsonResponse(http.StatusOK, projectStatus)

}
