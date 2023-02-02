package v1_deployment

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	deploymentModels "go-deploy/models/deployment"
	"go-deploy/models/dto"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/validator"
	"go-deploy/service/deployment_service"
	"go-deploy/service/user_info_service"
	"net/http"
	"strconv"
)

func getURL(deployment *deploymentModels.Deployment) string {
	var baseURL string

	if len(deployment.Subsystems.Npm.ProxyHost.DomainNames) > 0 {
		baseURL = deployment.Subsystems.Npm.ProxyHost.DomainNames[0]
	} else {
		baseURL = "pending"
	}

	return baseURL
}

func getAll(_ string, context *app.ClientContext) {
	deployments, _ := deployment_service.GetAll()

	dtoDeployments := make([]dto.DeploymentRead, len(deployments))
	for i, deployment := range deployments {
		_, statusMsg, _ := deployment_service.GetStatusByID(deployment.ID)

		dtoDeployments[i] = deployment.ToDTO(statusMsg, getURL(&deployment))
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
		_, statusMsg, _ := deployment_service.GetStatusByID(deployment.ID)
		dtoDeployments[i] = deployment.ToDTO(statusMsg, getURL(&deployment))
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

	deployment, _ := deployment_service.GetByFullID(userID, deploymentID)

	if deployment == nil {
		context.NotFound()
		return
	}

	_, statusMsg, _ := deployment_service.GetStatusByID(deployment.ID)
	context.JSONResponse(200, deployment.ToDTO(statusMsg, getURL(deployment)))
}

func Create(c *gin.Context) {
	context := app.NewContext(c)

	bodyRules := validator.MapData{
		"name": []string{
			"required",
			"regex:^[a-zA-Z]([a-zA-Z0-9-]*[a-zA-Z0-9])?([a-zA-Z]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$",
			"min:3",
			"max:30",
		},
	}

	messages := validator.MapData{
		"name": []string{
			"required:Name is required",
			"regexp:Name must follow RFC 1035 and must not include any dots",
			"min:Name must be between 3-30 characters",
			"max:Name must be between 3-30 characters",
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

	userInfo, err := user_info_service.GetByToken(token)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if userInfo.DeploymentQuota == 0 {
		context.ErrorResponse(http.StatusUnauthorized, status_codes.Error, "User is not allowed to create deployments")
		return
	}

	userID := token.Sub

	exists, deployment, err := deployment_service.Exists(requestBody.Name)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if exists {
		if deployment.OwnerID != userID {
			context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceAlreadyExists, "Resource already exists")
			return
		}
		if deployment.BeingDeleted {
			context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingDeleted, "Resource is currently being deleted")
			return
		}
		deployment_service.Create(deployment.ID, requestBody.Name, userID)
		context.JSONResponse(http.StatusCreated, dto.DeploymentCreated{ID: deployment.ID})
		return
	}

	deploymentCount, err := deployment_service.GetCount(userID)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if deploymentCount >= userInfo.DeploymentQuota {
		context.ErrorResponse(http.StatusUnauthorized, status_codes.Error, fmt.Sprintf("User is not allowed to create more than %d deployments", userInfo.DeploymentQuota))
		return
	}

	deploymentID := uuid.New().String()
	deployment_service.Create(deploymentID, requestBody.Name, userID)
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
	userID := token.Sub
	deploymentID := context.GinContext.Param("deploymentId")

	currentDeployment, err := deployment_service.GetByFullID(userID, deploymentID)
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
