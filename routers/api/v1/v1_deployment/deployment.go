package v1_deployment

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto/v1/body"
	"go-deploy/models/dto/v1/query"
	"go-deploy/models/dto/v1/uri"
	deploymentModels "go-deploy/models/sys/deployment"
	jobModels "go-deploy/models/sys/job"
	zoneModels "go-deploy/models/sys/zone"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	v12 "go-deploy/service/v1/common"
	"go-deploy/service/v1/deployments/opts"
	teamOpts "go-deploy/service/v1/teams/opts"
)

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

	var userID string
	if requestQuery.UserID != nil {
		userID = *requestQuery.UserID
	} else if !requestQuery.All {
		userID = auth.UserID
	}

	deployV1 := service.V1(auth)

	deployments, err := deployV1.Deployments().List(opts.ListOpts{
		UserID:     &userID,
		Pagination: v12.GetOrDefaultPagination(requestQuery.Pagination),
		Shared:     true,
	})
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
		teamIDs, _ := deployV1.Teams().ListIDs(teamOpts.ListOpts{ResourceID: deployment.ID})
		dtoDeployments[i] = deployment.ToDTO(deployV1.SMs().GetURL(deployment.OwnerID), teamIDs)
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

	deployV1 := service.V1(auth)

	deployment, err := deployV1.Deployments().Get(requestURI.DeploymentID, opts.GetOpts{
		Shared: true,
	})
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	if deployment == nil {
		context.NotFound("Deployment not found")
		return
	}

	teamIDs, _ := deployV1.Teams().ListIDs(teamOpts.ListOpts{ResourceID: deployment.ID})
	context.Ok(deployment.ToDTO(deployV1.SMs().GetURL(deployment.OwnerID), teamIDs))
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

	deployV1 := service.V1(auth)

	doesNotAlreadyExists, err := deployV1.Deployments().NameAvailable(requestBody.Name)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if !doesNotAlreadyExists {
		context.UserError("Deployment already exists")
		return
	}

	if effectiveRole.Quotas.Deployments <= 0 {
		context.Forbidden("User is not allowed to create deployments")
		return
	}

	if requestBody.Zone != nil {
		zone := deployV1.Zones().Get(*requestBody.Zone, zoneModels.TypeDeployment)
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
		validGhToken, reason, err := deployV1.Deployments().ValidGitHubToken(requestBody.GitHub.Token)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if !validGhToken {
			context.Unauthorized(reason)
			return
		}

		validGitHubRepository, reason, err := deployV1.Deployments().ValidGitHubRepository(requestBody.GitHub.Token, requestBody.GitHub.RepositoryID)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if !validGitHubRepository {
			context.Unauthorized(reason)
			return
		}
	}

	err = deployV1.Deployments().CheckQuota("", &opts.QuotaOptions{Create: &requestBody})
	if err != nil {
		var quotaExceededErr sErrors.QuotaExceededError
		if errors.As(err, &quotaExceededErr) {
			context.Forbidden(quotaExceededErr.Error())
			return
		}

		context.ServerError(err, v1.InternalError)
		return
	}

	deploymentID := uuid.New().String()
	jobID := uuid.New().String()
	err = deployV1.Jobs().Create(jobID, auth.UserID, jobModels.TypeCreateDeployment, map[string]interface{}{
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

	deployV1 := service.V1(auth)

	currentDeployment, err := deployV1.Deployments().Get(requestURI.DeploymentID)
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

	err = deployV1.Deployments().StartActivity(requestURI.DeploymentID, deploymentModels.ActivityBeingDeleted)
	if err != nil {
		var failedToStartActivityErr sErrors.FailedToStartActivityError
		if errors.As(err, &failedToStartActivityErr) {
			context.Locked(failedToStartActivityErr.Error())
			return
		}

		if errors.Is(err, sErrors.DeploymentNotFoundErr) {
			context.NotFound("Deployment not found")
			return
		}

		context.ServerError(err, v1.InternalError)
		return
	}

	jobID := uuid.NewString()
	err = deployV1.Jobs().Create(jobID, auth.UserID, jobModels.TypeDeleteDeployment, map[string]interface{}{
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
// @Router /deployments/{deploymentId} [post]
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

	deployV1 := service.V1(auth)

	var deployment *deploymentModels.Deployment
	if requestBody.TransferCode != nil {
		deployment, err = deployV1.Deployments().Get(requestURI.DeploymentID, opts.GetOpts{TransferCode: *requestBody.TransferCode})
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if requestBody.OwnerID == nil {
			requestBody.OwnerID = &auth.UserID
		}
	} else {
		deployment, err = deployV1.Deployments().Get(requestURI.DeploymentID, opts.GetOpts{Shared: true})
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}
	}

	if deployment == nil {
		context.NotFound("Deployment not found")
		return
	}

	if requestBody.OwnerID != nil {
		if *requestBody.OwnerID == "" {
			err = deployV1.Deployments().ClearUpdateOwner(requestURI.DeploymentID)
			if err != nil {
				if errors.Is(err, sErrors.DeploymentNotFoundErr) {
					context.NotFound("Deployment not found")
					return
				}

				context.ServerError(err, v1.InternalError)
				return
			}

			context.Ok(body.DeploymentUpdated{
				ID: deployment.ID,
			})
			return
		}

		if *requestBody.OwnerID == deployment.OwnerID {
			context.UserError("Owner already set")
			return
		}

		exists, err := deployV1.Users().Exists(*requestBody.OwnerID)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if !exists {
			context.UserError("User not found")
			return
		}

		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		jobID, err := deployV1.Deployments().UpdateOwnerSetup(requestURI.DeploymentID, &body.DeploymentUpdateOwner{
			NewOwnerID:   *requestBody.OwnerID,
			OldOwnerID:   deployment.OwnerID,
			TransferCode: requestBody.TransferCode,
		})

		if err != nil {
			if errors.Is(err, sErrors.DeploymentNotFoundErr) {
				context.NotFound("Deployment not found")
				return
			}

			if errors.Is(err, sErrors.InvalidTransferCodeErr) {
				context.Forbidden("Bad transfer code")
				return
			}

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
		available, err := deployV1.Deployments().NameAvailable(*requestBody.Name)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if !available {
			context.UserError("Name already taken")
			return
		}
	}

	err = deployV1.Deployments().CheckQuota(requestURI.DeploymentID, &opts.QuotaOptions{Update: &requestBody})
	if err != nil {
		var quotaExceededErr sErrors.QuotaExceededError
		if errors.As(err, &quotaExceededErr) {
			context.Forbidden(quotaExceededErr.Error())
			return
		}

		context.ServerError(err, v1.InternalError)
		return
	}

	canUpdate, reason := deployV1.Deployments().CanAddActivity(requestURI.DeploymentID, deploymentModels.ActivityUpdating)
	if !canUpdate {
		context.Locked(reason)
		return
	}

	jobID := uuid.New().String()
	err = deployV1.Jobs().Create(jobID, auth.UserID, jobModels.TypeUpdateDeployment, map[string]interface{}{
		"id":     deployment.ID,
		"params": requestBody,
	})

	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(body.DeploymentUpdated{
		ID:    deployment.ID,
		JobID: &jobID,
	})
}
