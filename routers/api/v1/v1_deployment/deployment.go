package v1_deployment

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	deploymentModels "go-deploy/models/deployment"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	jobModel "go-deploy/models/job"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/deployment_service"
	"go-deploy/service/job_service"
	"go-deploy/service/user_service"
	"net/http"
)

func getURL(deployment *deploymentModels.Deployment) string {
	var url string

	if len(deployment.Subsystems.Npm.ProxyHost.DomainNames) > 0 {
		url = deployment.Subsystems.Npm.ProxyHost.DomainNames[0]
	} else {
		url = "notset"
	}

	return url
}

func getAll(_ string, context *app.ClientContext) {
	deployments, _ := deployment_service.GetAll()

	dtoDeployments := make([]body.DeploymentRead, len(deployments))
	for i, deployment := range deployments {
		dtoDeployments[i] = deployment.ToDTO(getURL(&deployment))
	}

	context.JSONResponse(http.StatusOK, dtoDeployments)
}

func GetList(c *gin.Context) {
	context := app.NewContext(c)

	var requestQuery query.DeploymentList
	if err := context.GinContext.BindQuery(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(&requestQuery, err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err.Error()))
		return
	}

	if requestQuery.WantAll && auth.IsAdmin {
		getAll(auth.UserID, &context)
		return
	}

	deployments, _ := deployment_service.GetByOwnerID(auth.UserID)
	if deployments == nil {
		context.JSONResponse(200, []interface{}{})
		return
	}

	dtoDeployments := make([]body.DeploymentRead, len(deployments))
	for i, deployment := range deployments {
		dtoDeployments[i] = deployment.ToDTO(getURL(&deployment))
	}

	context.JSONResponse(200, dtoDeployments)
}

func Get(c *gin.Context) {
	context := app.NewContext(c)

	var requestURI uri.DeploymentGet
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(&requestURI, err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	deployment, err := deployment_service.GetByID(auth.UserID, requestURI.DeploymentID, auth.IsAdmin)

	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if deployment == nil {
		context.NotFound()
		return
	}

	context.JSONResponse(200, deployment.ToDTO(getURL(deployment)))
}

func Create(c *gin.Context) {
	context := app.NewContext(c)

	var requestBody body.DeploymentCreate
	if err := context.GinContext.BindJSON(&requestBody); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(&requestBody, err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	userInfo, err := user_service.GetOrCreate(auth.JwtToken)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if userInfo.DeploymentQuota == 0 {
		context.ErrorResponse(http.StatusUnauthorized, status_codes.Error, "User is not allowed to create deployments")
		return
	}

	exists, deployment, err := deployment_service.Exists(requestBody.Name)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if exists {
		if deployment.OwnerID != auth.UserID {
			context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceNotCreated, "Resource already exists")
			return
		}
		if deployment.BeingDeleted {
			context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingDeleted, "Resource is currently being deleted")
			return
		}
		jobID := uuid.New().String()
		err = job_service.Create(jobID, auth.UserID, jobModel.TypeCreateDeployment, map[string]interface{}{
			"id":      deployment.ID,
			"name":    requestBody.Name,
			"ownerId": auth.UserID,
		})
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
			return
		}

		context.JSONResponse(http.StatusCreated, body.DeploymentCreated{
			ID:    deployment.ID,
			JobID: jobID,
		})
		return
	}

	deploymentCount, err := deployment_service.GetCount(auth.UserID)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if deploymentCount >= userInfo.DeploymentQuota {
		context.ErrorResponse(http.StatusUnauthorized, status_codes.Error, fmt.Sprintf("User is not allowed to create more than %d deployments", userInfo.DeploymentQuota))
		return
	}

	deploymentID := uuid.New().String()
	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeCreateDeployment, map[string]interface{}{
		"id":      deploymentID,
		"name":    requestBody.Name,
		"ownerId": auth.UserID,
	})

	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	context.JSONResponse(http.StatusCreated, body.DeploymentCreated{
		ID:    deploymentID,
		JobID: jobID,
	})
}

func Delete(c *gin.Context) {
	context := app.NewContext(c)

	var requestURI uri.DeploymentDelete
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(&requestURI, err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	currentDeployment, err := deployment_service.GetByID(auth.UserID, requestURI.DeploymentID, auth.IsAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if currentDeployment == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, "Resource not found")
		return
	}

	if currentDeployment.BeingCreated {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingCreated, "Resource is currently being created")
		return
	}

	if !currentDeployment.BeingDeleted {
		_ = deployment_service.MarkBeingDeleted(currentDeployment.ID)
	}

	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeDeleteDeployment, map[string]interface{}{
		"name": currentDeployment.Name,
	})
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	context.JSONResponse(http.StatusOK, body.DeploymentDeleted{
		ID:    currentDeployment.ID,
		JobID: jobID,
	})
}
