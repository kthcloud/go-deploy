package v1_deployment

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	deploymentModels "go-deploy/models/sys/deployment"
	jobModel "go-deploy/models/sys/job"
	vmModel "go-deploy/models/sys/vm"
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

// GetList
// @Summary Get list of deployments
// @Description Get list of deployments
// @BasePath /api/v1
// @Tags Deployment
// @Accept  json
// @Produce  json
// @Param Authorization header string true "With the bearer started"
// @Param wantAll query bool false "Get all deployments"
// @Success 200 {array} body.DeploymentRead
// @Failure 400 {object} app.ErrorResponse
// @Failure 500 {object} app.ErrorResponse
// @Router /api/v1/deployments [get]
func GetList(c *gin.Context) {
	context := app.NewContext(c)

	var requestQuery query.DeploymentList
	if err := context.GinContext.Bind(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err.Error()))
		return
	}

	if requestQuery.WantAll && auth.IsAdmin() {
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

// Get
// @Summary Get deployment by id
// @Description Get deployment by id
// @BasePath /api/v1
// @Tags Deployment
// @Accept  json
// @Produce  json
// @Param Authorization header string true "With the bearer started"
// @Param deploymentId path string true "Deployment ID"
// @Success 200 {object} body.DeploymentRead
// @Failure 400 {object} app.ErrorResponse
// @Failure 500 {object} app.ErrorResponse
// @Router /api/v1/deployments/{deployment_id} [get]
func Get(c *gin.Context) {
	context := app.NewContext(c)

	var requestURI uri.DeploymentGet
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	deployment, err := deployment_service.GetByID(auth.UserID, requestURI.DeploymentID, auth.IsAdmin())

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

// Create
// @Summary Create deployment
// @Description Create deployment
// @BasePath /api/v1
// @Tags Deployment
// @Accept  json
// @Produce  json
// @Param Authorization header string true "With the bearer started"
// @Param body body body.DeploymentCreate true "Deployment body"
// @Success 200 {object} body.DeploymentRead
// @Failure 400 {object} app.ErrorResponse
// @Failure 500 {object} app.ErrorResponse
// @Router /api/v1/deployments [post]
func Create(c *gin.Context) {
	context := app.NewContext(c)

	var requestBody body.DeploymentCreate
	if err := context.GinContext.BindJSON(&requestBody); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	user, err := user_service.GetOrCreate(auth.UserID, auth.JwtToken.PreferredUsername, auth.Roles)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get user: %s", err))
		return
	}

	quota, err := user_service.GetQuotaByUserID(user.ID)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get user quota: %s", err))
		return
	}

	if quota.Deployments <= 0 {
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

		started, reason, err := deployment_service.StartActivity(deployment.ID, vmModel.ActivityBeingCreated)
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to start activity: %s", err))
			return
		}

		if !started {
			context.ErrorResponse(http.StatusLocked, status_codes.ResourceNotReady, reason)
			return
		}

		jobID := uuid.New().String()
		err = job_service.Create(jobID, auth.UserID, jobModel.TypeCreateDeployment, map[string]interface{}{
			"id":      deployment.ID,
			"ownerId": auth.UserID,
			"params":  requestBody,
		})
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to create job: %s", err))
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

	if deploymentCount >= quota.Deployments {
		context.ErrorResponse(http.StatusUnauthorized, status_codes.Error, fmt.Sprintf("User is not allowed to create more than %d deployments", quota.Deployments))
		return
	}

	deploymentID := uuid.New().String()
	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeCreateDeployment, map[string]interface{}{
		"id":      deploymentID,
		"ownerId": auth.UserID,
		"params":  requestBody,
	})

	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to create job: %s", err))
		return
	}

	context.JSONResponse(http.StatusCreated, body.DeploymentCreated{
		ID:    deploymentID,
		JobID: jobID,
	})
}

// Delete
// @Summary Delete deployment
// @Description Delete deployment
// @BasePath /api/v1
// @Tags Deployment
// @Accept  json
// @Produce  json
// @Param Authorization header string true "With the bearer started"
// @Param deploymentId path string true "Deployment ID"
// @Success 200 {object} body.DeploymentCreated
// @Failure 400 {object} app.ErrorResponse
// @Failure 500 {object} app.ErrorResponse
// @Router /api/v1/deployments/{deploymentId} [delete]
func Delete(c *gin.Context) {
	context := app.NewContext(c)

	var requestURI uri.DeploymentDelete
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	currentDeployment, err := deployment_service.GetByID(auth.UserID, requestURI.DeploymentID, auth.IsAdmin())
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if currentDeployment == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, "Resource not found")
		return
	}

	added, reason, err := deployment_service.StartActivity(currentDeployment.ID, deploymentModels.ActivityBeingDeleted)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to start activity: %s", err))
		return
	}

	if !added {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceNotUpdated, fmt.Sprintf("Could not transition to delete state: %s", reason))
		return
	}

	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeDeleteDeployment, map[string]interface{}{
		"name": currentDeployment.Name,
	})
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to create job: %s", err))
		return
	}

	context.JSONResponse(http.StatusOK, body.DeploymentDeleted{
		ID:    currentDeployment.ID,
		JobID: jobID,
	})
}

// Update
// @Summary Update deployment
// @Description Update deployment
// @BasePath /api/v1
// @Tags Deployment
// @Accept  json
// @Produce  json
// @Param Authorization header string true "With the bearer started"
// @Param deploymentId path string true "Deployment ID"
// @Param body body body.DeploymentUpdate true "Deployment update"
// @Success 200 {object} body.DeploymentUpdated
// @Failure 400 {object} app.ErrorResponse
// @Failure 500 {object} app.ErrorResponse
// @Router /api/v1/deployments/{deploymentId} [put]
func Update(c *gin.Context) {
	context := app.NewContext(c)

	var requestURI uri.DeploymentUpdate
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	var requestBody body.DeploymentUpdate
	if err := context.GinContext.BindJSON(&requestBody); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	current, err := deployment_service.GetByID(auth.UserID, requestURI.DeploymentID, auth.IsAdmin())
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, fmt.Sprintf("Failed to get vm: %s", err))
		return
	}

	if current == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("Deployment with id %s not found", requestURI.DeploymentID))
		return
	}

	if current.BeingCreated() {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingCreated, "Resource is currently being created")
		return
	}

	if current.BeingDeleted() {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingDeleted, "Resource is currently being deleted")
		return
	}

	if current.OwnerID != auth.UserID && !auth.IsAdmin() {
		context.ErrorResponse(http.StatusUnauthorized, status_codes.Error, "User is not allowed to update this resource")
		return
	}

	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeUpdateDeployment, map[string]interface{}{
		"name":   current.Name,
		"update": requestBody,
	})

	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to create job: %s", err))
		return
	}

	context.JSONResponse(http.StatusOK, body.DeploymentUpdated{
		ID:    current.ID,
		JobID: jobID,
	})

}
