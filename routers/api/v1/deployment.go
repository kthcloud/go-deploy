package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/dto/v1/body"
	"go-deploy/dto/v1/query"
	"go-deploy/dto/v1/uri"
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/sys"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/v1/deployments/opts"
	teamOpts "go-deploy/service/v1/teams/opts"
	v12 "go-deploy/service/v1/utils"
)

// GetDeployment
// @Summary Get deployment
// @Description Get deployment
// @Tags Deployment
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param deploymentId path string true "Deployment ID"
// @Success 200 {object} body.DeploymentRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v1/deployments/{deployment_id} [get]
func GetDeployment(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.DeploymentGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)

	deployment, err := deployV1.Deployments().Get(requestURI.DeploymentID, opts.GetOpts{
		Shared: true,
	})
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	if deployment == nil {
		context.NotFound("Deployment not found")
		return
	}

	teamIDs, _ := deployV1.Teams().ListIDs(teamOpts.ListOpts{ResourceID: deployment.ID})
	context.Ok(deployment.ToDTO(deployV1.SMs().GetURL(deployment.OwnerID), teamIDs))
}

// ListDeployments
// @Summary List deployments
// @Description List deployments
// @Tags Deployment
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param all query bool false "GetDeployment all"
// @Param userId query string false "Filter by user id"
// @Param shared query bool false "Include shared"
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {array} body.DeploymentRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v1/deployments [get]
func ListDeployments(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.DeploymentList
	if err := context.GinContext.Bind(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
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
		context.ServerError(err, AuthInfoNotAvailableErr)
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

// CreateDeployment
// @Summary Create deployment
// @Description Create deployment
// @Tags Deployment
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param body body body.DeploymentCreate true "Deployment body"
// @Success 200 {object} body.DeploymentRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v1/deployments [post]
func CreateDeployment(c *gin.Context) {
	context := sys.NewContext(c)

	var requestBody body.DeploymentCreate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	effectiveRole := auth.GetEffectiveRole()
	if effectiveRole == nil {
		context.ServerError(err, InternalError)
		return
	}

	deployV1 := service.V1(auth)

	doesNotAlreadyExists, err := deployV1.Deployments().NameAvailable(requestBody.Name)
	if err != nil {
		context.ServerError(err, InternalError)
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
		zone := deployV1.Zones().Get(*requestBody.Zone, model.ZoneTypeDeployment)
		if zone == nil {
			context.NotFound("Zone not found")
			return
		}
	}

	if requestBody.CustomDomain != nil && !effectiveRole.Permissions.UseCustomDomains {
		context.Forbidden("User is not allowed to use custom domains")
		return
	}

	err = deployV1.Deployments().CheckQuota("", &opts.QuotaOptions{Create: &requestBody})
	if err != nil {
		var quotaExceededErr sErrors.QuotaExceededError
		if errors.As(err, &quotaExceededErr) {
			context.Forbidden(quotaExceededErr.Error())
			return
		}

		context.ServerError(err, InternalError)
		return
	}

	deploymentID := uuid.New().String()
	jobID := uuid.New().String()
	err = deployV1.Jobs().Create(jobID, auth.UserID, model.JobCreateDeployment, version.V1, map[string]interface{}{
		"id":      deploymentID,
		"ownerId": auth.UserID,
		"params":  requestBody,
	})

	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	context.Ok(body.DeploymentCreated{
		ID:    deploymentID,
		JobID: jobID,
	})
}

// DeleteDeployment
// @Summary Delete deployment
// @Description Delete deployment
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
// @Router /v1/deployments/{deploymentId} [delete]
func DeleteDeployment(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.DeploymentDelete
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)

	currentDeployment, err := deployV1.Deployments().Get(requestURI.DeploymentID)
	if err != nil {
		context.ServerError(err, InternalError)
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

	err = deployV1.Deployments().StartActivity(requestURI.DeploymentID, model.ActivityBeingDeleted)
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

		context.ServerError(err, InternalError)
		return
	}

	jobID := uuid.NewString()
	err = deployV1.Jobs().Create(jobID, auth.UserID, model.JobDeleteDeployment, version.V1, map[string]interface{}{
		"id": currentDeployment.ID,
	})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	context.Ok(body.DeploymentDeleted{
		ID:    currentDeployment.ID,
		JobID: jobID,
	})
}

// UpdateDeployment
// @Summary Update deployment
// @Description Update deployment
// @Tags Deployment
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param deploymentId path string true "Deployment ID"
// @Param body body body.DeploymentUpdate true "Deployment update"
// @Success 200 {object} body.DeploymentUpdated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v1/deployments/{deploymentId} [post]
func UpdateDeployment(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.DeploymentUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var requestBody body.DeploymentUpdate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)

	var deployment *model.Deployment
	if requestBody.TransferCode != nil {
		deployment, err = deployV1.Deployments().Get(requestURI.DeploymentID, opts.GetOpts{TransferCode: *requestBody.TransferCode})
		if err != nil {
			context.ServerError(err, InternalError)
			return
		}

		if requestBody.OwnerID == nil {
			requestBody.OwnerID = &auth.UserID
		}
	} else {
		deployment, err = deployV1.Deployments().Get(requestURI.DeploymentID, opts.GetOpts{Shared: true})
		if err != nil {
			context.ServerError(err, InternalError)
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

				context.ServerError(err, InternalError)
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
			context.ServerError(err, InternalError)
			return
		}

		if !exists {
			context.UserError("User not found")
			return
		}

		if err != nil {
			context.ServerError(err, InternalError)
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

			context.ServerError(err, InternalError)
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
			context.ServerError(err, InternalError)
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

		context.ServerError(err, InternalError)
		return
	}

	canUpdate, reason := deployV1.Deployments().CanAddActivity(requestURI.DeploymentID, model.ActivityUpdating)
	if !canUpdate {
		context.Locked(reason)
		return
	}

	jobID := uuid.New().String()
	err = deployV1.Jobs().Create(jobID, auth.UserID, model.JobUpdateDeployment, version.V1, map[string]interface{}{
		"id":     deployment.ID,
		"params": requestBody,
	})

	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	context.Ok(body.DeploymentUpdated{
		ID:    deployment.ID,
		JobID: &jobID,
	})
}
