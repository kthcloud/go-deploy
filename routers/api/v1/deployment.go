package v1

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

func getAllDeployments(context *app.ClientContext) {
	deployments, _ := deployment_service.GetAll()

	dtoDeployments := make([]dto.Deployment, len(deployments))
	for i, deployment := range deployments {
		dtoDeployments[i] = deployment.ToDto()
	}

	context.JSONResponse(http.StatusOK, dtoDeployments)
}

func GetDeployments(c *gin.Context) {
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
		getAllDeployments(&context)
		return
	}

	deployments, _ := deployment_service.GetByOwner(userID)
	if deployments == nil {
		context.JSONResponse(200, []interface{}{})
		return
	}

	dtoDeployments := make([]dto.Deployment, len(deployments))
	for i, deployment := range deployments {
		dtoDeployments[i] = deployment.ToDto()
	}

	context.JSONResponse(200, dtoDeployments)
}

func GetDeployment(c *gin.Context) {
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

	deployment, _ := deployment_service.Get(userId, deploymentID)

	if deployment == nil {
		context.NotFound()
		return
	}

	context.JSONResponse(200, deployment.ToDto())
}

func CreateDeployment(c *gin.Context) {
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
			"required:Deployment name is required",
			"regexp:Deployment name must be all lowercase",
			"min:Deployment name must be between 3-10 characters",
			"max:Deployment name must be between 3-10 characters",
		},
	}

	var requestBody dto.DeploymentReq
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
		context.ErrorResponse(http.StatusInternalServerError, status_codes.DeploymentValidationFailed, "Failed to validate deployment")
		return
	}

	if exists {
		if deployment.Owner != userId {
			context.ErrorResponse(http.StatusBadRequest, status_codes.DeploymentAlreadyExists, "Deployment already exists")
			return
		}
		if deployment.BeingDeleted {
			context.ErrorResponse(http.StatusLocked, status_codes.DeploymentBeingDeleted, "Deployment is currently being deleted")
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

func DeleteDeployment(c *gin.Context) {
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

	currentDeployment, err := deployment_service.Get(userId, deploymentID)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.DeploymentValidationFailed, "Failed to validate deployment")
		return
	}

	if currentDeployment == nil {
		context.NotFound()
		return
	}

	if currentDeployment.BeingCreated {
		context.ErrorResponse(http.StatusLocked, status_codes.DeploymentBeingCreated, "Deployment is currently being created")
		return
	}

	if !currentDeployment.BeingDeleted {
		_ = deployment_service.MarkBeingDeleted(currentDeployment.ID)
	}

	deployment_service.Delete(currentDeployment.Name)

	context.OkDeleted()
}
