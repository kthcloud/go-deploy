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

func GetStatus(c *gin.Context) {
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
	}
	deploymentID := context.GinContext.Param("deploymentId")
	userId := token.Sub

	statusCode, deploymentStatus, _ := deployment_service.GetStatusByID(userId, deploymentID)
	if deploymentStatus == nil {
		context.NotFound()
		return
	}

	if statusCode == status_codes.DeploymentNotFound {
		context.JSONResponse(http.StatusNotFound, deploymentStatus)
		return
	}

	context.JSONResponse(http.StatusOK, deploymentStatus)

}
