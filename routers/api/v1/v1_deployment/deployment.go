package v1_deployment

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/validator"
	"go-deploy/service/deployment_service"
	"net/http"
	"strconv"
)

func getAll(userID string, context *app.ClientContext) {
	deployments, _ := deployment_service.GetAll()

	dtoDeployments := make([]dto.DeploymentRead, len(deployments))
	for i, deployment := range deployments {
		_, statusMsg, _ := deployment_service.GetStatusByID(userID, deployment.ID)
		dtoDeployments[i] = deployment.ToDto(statusMsg)
	}

	context.JSONResponse(http.StatusOK, dtoDeployments)
}

func GetMany(c *gin.Context) {
	context := app.NewContext(c)

	rules := validator.MapData{
		"all": []string{"bool"},
	}

	validationErrors := context.ValidateQueryParams(&rules)
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

	// might want to check if userID is allowed to get all...
	wantAll, _ := strconv.ParseBool(context.GinContext.Query("all"))
	if wantAll {
		getAll(userID, &context)
		return
	}

	deployments, _ := deployment_service.GetByOwnerID(userID)
	if deployments == nil {
		context.JSONResponse(200, []interface{}{})
		return
	}

	dtoDeployments := make([]dto.DeploymentRead, len(deployments))
	for i, deployment := range deployments {
		_, statusMsg, _ := deployment_service.GetStatusByID(userID, deployment.ID)
		dtoDeployments[i] = deployment.ToDto(statusMsg)
	}

	context.JSONResponse(200, dtoDeployments)
}

func Get(c *gin.Context) {
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
	userID := token.Sub

	deployment, _ := deployment_service.GetByID(userID, deploymentID)

	if deployment == nil {
		context.NotFound()
		return
	}

	_, statusMsg, _ := deployment_service.GetStatusByID(userID, deployment.ID)
	context.JSONResponse(200, deployment.ToDto(statusMsg))
}

func Create(c *gin.Context) {
	context := app.NewContext(c)

	bodyRules := validator.MapData{
		"name": []string{
			"required",
			"regex:^[a-zA-Z]+$",
			"min:3",
			"max:10",
		},
	}

	messages := validator.MapData{
		"name": []string{
			"required:Name is required",
			"regexp:Name must be all lowercase",
			"min:Name must be between 3-10 characters",
			"max:Name must be between 3-10 characters",
		},
	}

	var requestBody dto.DeploymentCreate
	validationErrors := context.ValidateJSONCustomMessages(&bodyRules, &messages, &requestBody)
	if len(validationErrors) > 0 {
		context.ResponseValidationError(validationErrors)
		return
	}

	token, err := context.GetKeycloakToken()
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}
	userId := token.Sub

	exists, deployment, err := deployment_service.Exists(requestBody.Name)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if exists {
		if deployment.Owner != userId {
			context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceAlreadyExists, "Resource already exists")
			return
		}
		if deployment.BeingDeleted {
			context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingDeleted, "Resource is currently being deleted")
			return
		}
		deployment_service.Create(deployment.ID, requestBody.Name, userId)
		context.JSONResponse(http.StatusCreated, dto.DeploymentCreated{ID: deployment.ID})
		return
	}

	deploymentID := uuid.New().String()
	deployment_service.Create(deploymentID, requestBody.Name, userId)
	context.JSONResponse(http.StatusCreated, dto.DeploymentCreated{ID: deploymentID})
}

func Delete(c *gin.Context) {
	context := app.NewContext(c)

	rules := validator.MapData{
		"deploymentId": []string{
			"required",
			"uuid_v4",
		},
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
	userId := token.Sub
	deploymentID := context.GinContext.Param("deploymentId")

	currentDeployment, err := deployment_service.GetByID(userId, deploymentID)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if currentDeployment == nil {
		context.NotFound()
		return
	}

	if currentDeployment.BeingCreated {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingCreated, "Resource is currently being created")
		return
	}

	if !currentDeployment.BeingDeleted {
		_ = deployment_service.MarkBeingDeleted(currentDeployment.ID)
	}

	deployment_service.Delete(currentDeployment.Name)

	context.OkDeleted()
}
