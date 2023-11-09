package v1_deployment

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	deploymentModels "go-deploy/models/sys/deployment"
	jobModel "go-deploy/models/sys/job"
	zoneModel "go-deploy/models/sys/zone"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	"go-deploy/service/deployment_service"
	"go-deploy/service/job_service"
	"go-deploy/service/storage_manager_service"
	"go-deploy/service/user_service"
	"go-deploy/service/zone_service"
)

func getStorageManagerURL(userID string, auth *service.AuthInfo) *string {
	storageManager, err := storage_manager_service.GetStorageManagerByOwnerIdAuth(userID, auth)
	if err != nil {
		return nil
	}

	if storageManager == nil {
		return nil
	}

	return storageManager.GetURL()
}

// List
// @Summary Get list of deployments
// @Description Get list of deployments
// @BasePath /api/v1
// @Tags Deployment
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param all query bool false "Get all"
// @Param userId query string false "Filter by user id"
// @Param shared query bool false "Include shared"
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {array} body.DeploymentRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /deployments [get]
func List(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.DeploymentList
	if err := context.GinContext.Bind(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	deployments, err := deployment_service.ListAuth(requestQuery.All, requestQuery.UserID, requestQuery.Shared, auth, &requestQuery.Pagination)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	if deployments == nil {
		context.JSONResponse(200, []interface{}{})
		return
	}

	dtoDeployments := make([]body.DeploymentRead, len(deployments))
	for i, deployment := range deployments {
		var storageManagerURL *string
		if mainApp := deployment.GetMainApp(); mainApp != nil && len(mainApp.Volumes) > 0 {
			storageManagerURL = getStorageManagerURL(deployment.OwnerID, auth)
		}

		dtoDeployments[i] = deployment.ToDTO(storageManagerURL)
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
// @Param Authorization header string true "Bearer token"
// @Param deploymentId path string true "Deployment ID"
// @Success 200 {object} body.DeploymentRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /deployments/{deployment_id} [get]
func Get(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.DeploymentGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	deployment, err := deployment_service.GetByIdAuth(requestURI.DeploymentID, auth)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	if deployment == nil {
		context.NotFound("Deployment not found")
		return
	}

	var storageManagerURL *string
	if len(deployment.GetMainApp().Volumes) > 0 {
		storageManagerURL = getStorageManagerURL(deployment.OwnerID, auth)
	}

	context.Ok(deployment.ToDTO(storageManagerURL))
}

// Create
// @Summary Create deployment
// @Description Create deployment
// @BasePath /api/v1
// @Tags Deployment
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param body body body.DeploymentCreate true "Deployment body"
// @Success 200 {object} body.DeploymentRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /deployments [post]
func Create(c *gin.Context) {
	context := sys.NewContext(c)

	var requestBody body.DeploymentCreate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	effectiveRole := auth.GetEffectiveRole()
	if effectiveRole == nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	deployment, err := deployment_service.GetByName(requestBody.Name)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if deployment != nil {
		context.UserError("Deployment already exists")
		return
	}

	if effectiveRole.Quotas.Deployments <= 0 {
		context.Forbidden("User is not allowed to create deployments")
		return
	}

	if requestBody.Zone != nil {
		zone := zone_service.GetZone(*requestBody.Zone, zoneModel.ZoneTypeDeployment)
		if zone == nil {
			context.NotFound("Zone not found")
			return
		}
	}

	if requestBody.CustomDomain != nil && !effectiveRole.Permissions.UseCustomDomains {
		context.Forbidden("User is not allowed to use custom domains")
		return
	}

	if requestBody.GitHub != nil {
		validGhToken, reason, err := deployment_service.ValidGitHubToken(requestBody.GitHub.Token)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if !validGhToken {
			context.Unauthorized(reason)
			return
		}

		validGitHubRepository, reason, err := deployment_service.ValidGitHubRepository(requestBody.GitHub.Token, requestBody.GitHub.RepositoryID)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if !validGitHubRepository {
			context.Unauthorized(reason)
			return
		}
	}

	ok, reason, err := deployment_service.CheckQuotaCreate(auth.UserID, &auth.GetEffectiveRole().Quotas, auth)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if !ok {
		context.Forbidden(reason)
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
		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(body.DeploymentCreated{
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
// @Param Authorization header string true "Bearer token"
// @Param deploymentId path string true "Deployment ID"
// @Success 200 {object} body.DeploymentCreated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /deployments/{deploymentId} [delete]
func Delete(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.DeploymentDelete
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	currentDeployment, err := deployment_service.GetByIdAuth(requestURI.DeploymentID, auth)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if currentDeployment == nil {
		context.NotFound("Deployment not found")
		return
	}

	if currentDeployment.OwnerID != auth.UserID && !auth.IsAdmin {
		context.Forbidden("Deployments can only be deleted by their owner")
		return
	}

	started, reason, err := deployment_service.StartActivity(currentDeployment.ID, deploymentModels.ActivityBeingDeleted)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if !started {
		context.Locked(reason)
		return
	}

	jobID := uuid.NewString()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeDeleteDeployment, map[string]interface{}{
		"id": currentDeployment.ID,
	})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(body.DeploymentDeleted{
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
// @Param Authorization header string true "Bearer token"
// @Param deploymentId path string true "Deployment ID"
// @Param body body body.DeploymentUpdate true "Deployment update"
// @Success 200 {object} body.DeploymentUpdated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /deployments/{deploymentId} [put]
func Update(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.DeploymentUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	var requestBody body.DeploymentUpdate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	deployment, err := deployment_service.GetByIdAuth(requestURI.DeploymentID, auth)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if deployment == nil {
		context.NotFound("Deployment not found")
		return
	}

	if requestBody.OwnerID != nil {
		if *requestBody.OwnerID == deployment.OwnerID {
			context.UserError("Owner already set")
			return
		}

		exists, err := user_service.Exists(*requestBody.OwnerID)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if !exists {
			context.UserError("User not found")
			return
		}

		jobID := uuid.New().String()
		err = job_service.Create(jobID, auth.UserID, jobModel.TypeUpdateDeploymentOwner, map[string]interface{}{
			"id": deployment.ID,
			"params": body.DeploymentUpdateOwner{
				NewOwnerID: *requestBody.OwnerID,
				OldOwnerID: deployment.OwnerID,
			},
		})

		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		context.Ok(body.DeploymentUpdated{
			ID:    deployment.ID,
			JobID: jobID,
		})

		return
	}

	if requestBody.Name != nil {
		available, err := deployment_service.NameAvailable(*requestBody.Name)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if !available {
			context.UserError("Name already taken")
			return
		}
	}

	canUpdate, reason := deployment_service.CanAddActivity(deployment.ID, deploymentModels.ActivityUpdating)
	if !canUpdate {
		context.Locked(reason)
		return
	}

	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeUpdateDeployment, map[string]interface{}{
		"id":     deployment.ID,
		"params": requestBody,
	})

	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(body.DeploymentUpdated{
		ID:    deployment.ID,
		JobID: jobID,
	})
}
