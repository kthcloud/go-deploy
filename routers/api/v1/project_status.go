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

func GetProjectStatus(c *gin.Context) {
	context := app.NewContext(c)

	rules := validator.MapData{
		"projectId": []string{"required", "uuid_v4"},
	}

	validationErrors := context.ValidateParams(&rules)

	if len(validationErrors) > 0 {
		context.ResponseValidationError(validationErrors)
		return
	}

	token, err := context.GetKeycloakToken()
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
	}
	projectId := context.GinContext.Param("projectId")
	userId := token.Sub

	statusCode, projectStatus, _ := project_service.GetStatusByID(userId, projectId)
	if projectStatus == nil {
		context.NotFound()
		return
	}

	if statusCode == status_codes.ProjectNotFound {
		context.JSONResponse(http.StatusNotFound, projectStatus)
		return
	}

	context.JSONResponse(http.StatusOK, projectStatus)

}
