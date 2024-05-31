package v2

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/dto/v2/body"
	"go-deploy/dto/v2/query"
	"go-deploy/dto/v2/uri"
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/config"
	"go-deploy/pkg/sys"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/v2/deployments/opts"
	teamOpts "go-deploy/service/v2/teams/opts"
	v12 "go-deploy/service/v2/utils"
	"strconv"
	"strings"
)

// GetDeployment
// @Summary Get deployment
// @Description Get deployment
// @Tags Deployment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param deploymentId path string true "Deployment ID"
// @Success 200 {object} body.DeploymentRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/deployments/{deploymentId} [get]
func GetDeployment(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.DeploymentGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var requestQuery query.DeploymentGet
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	deployV2 := service.V2(auth)

	var deployment *model.Deployment
	if requestQuery.MigrationCode != nil {
		deployment, err = deployV2.Deployments().Get(requestURI.DeploymentID, opts.GetOpts{MigrationCode: requestQuery.MigrationCode})
	} else {
		deployment, err = deployV2.Deployments().Get(requestURI.DeploymentID, opts.GetOpts{Shared: true})
	}

	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if deployment == nil {
		context.NotFound("Deployment not found")
		return
	}

	teamIDs, _ := deployV2.Teams().ListIDs(teamOpts.ListOpts{ResourceID: deployment.ID})
	context.Ok(deployment.ToDTO(deployV2.SMs().GetUrlByUserID(deployment.OwnerID), getDeploymentExternalPort(deployment.Zone), teamIDs))
}

// ListDeployments
// @Summary List deployments
// @Description List deployments
// @Tags Deployment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param all query bool false "List all"
// @Param userId query string false "Filter by user ID"
// @Param shared query bool false "Include shared"
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {array} body.DeploymentRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/deployments [get]
func ListDeployments(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.DeploymentList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
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
		userID = auth.User.ID
	}

	deployV2 := service.V2(auth)

	deployments, err := deployV2.Deployments().List(opts.ListOpts{
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
		teamIDs, _ := deployV2.Teams().ListIDs(teamOpts.ListOpts{ResourceID: deployment.ID})
		dtoDeployments[i] = deployment.ToDTO(deployV2.SMs().GetUrlByUserID(deployment.OwnerID), getDeploymentExternalPort(deployment.Zone), teamIDs)
	}

	context.JSONResponse(200, dtoDeployments)
}

// CreateDeployment
// @Summary Create deployment
// @Description Create deployment
// @Tags Deployment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body body.DeploymentCreate true "Deployment body"
// @Success 200 {object} body.DeploymentRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/deployments [post]
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

	deployV2 := service.V2(auth)

	doesNotAlreadyExists, err := deployV2.Deployments().NameAvailable(requestBody.Name)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if !doesNotAlreadyExists {
		context.UserError("Deployment already exists")
		return
	}

	if requestBody.Zone != nil {
		zone := deployV2.System().GetZone(*requestBody.Zone)
		if zone == nil {
			context.NotFound("Zone not found")
			return
		}

		if !zone.Enabled {
			context.Forbidden("Zone is disabled")
			return
		}

		if !deployV2.System().ZoneHasCapability(*requestBody.Zone, configModels.ZoneCapabilityDeployment) {
			context.Forbidden("Zone does not have deployment capability")
			return
		}
	}

	if requestBody.CustomDomain != nil && !effectiveRole.Permissions.UseCustomDomains {
		context.Forbidden("User is not allowed to use custom domains")
		return
	}

	err = deployV2.Deployments().CheckQuota("", &opts.QuotaOptions{Create: &requestBody})
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
	err = deployV2.Jobs().Create(jobID, auth.User.ID, model.JobCreateDeployment, version.V2, map[string]interface{}{
		"id":       deploymentID,
		"ownerId":  auth.User.ID,
		"params":   requestBody,
		"authInfo": auth,
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
// @Security ApiKeyAuth
// @Param deploymentId path string true "Deployment ID"
// @Success 200 {object} body.DeploymentCreated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/deployments/{deploymentId} [delete]
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

	deployV2 := service.V2(auth)

	currentDeployment, err := deployV2.Deployments().Get(requestURI.DeploymentID)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if currentDeployment == nil {
		context.NotFound("Deployment not found")
		return
	}

	if currentDeployment.OwnerID != auth.User.ID && !auth.User.IsAdmin {
		context.Forbidden("Deployments can only be deleted by their owner")
		return
	}

	err = deployV2.Deployments().StartActivity(requestURI.DeploymentID, model.ActivityBeingDeleted)
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
	err = deployV2.Jobs().Create(jobID, auth.User.ID, model.JobDeleteDeployment, version.V2, map[string]interface{}{
		"id":       currentDeployment.ID,
		"ownerId":  auth.User.ID,
		"authInfo": auth,
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
// @Security ApiKeyAuth
// @Param deploymentId path string true "Deployment ID"
// @Param body body body.DeploymentUpdate true "Deployment update"
// @Success 200 {object} body.DeploymentUpdated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/deployments/{deploymentId} [post]
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

	deployV2 := service.V2(auth)

	deployment, err := deployV2.Deployments().Get(requestURI.DeploymentID, opts.GetOpts{Shared: true})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}
	if deployment == nil {
		context.NotFound("Deployment not found")
		return
	}

	if requestBody.Name != nil {
		available, err := deployV2.Deployments().NameAvailable(*requestBody.Name)
		if err != nil {
			context.ServerError(err, InternalError)
			return
		}

		if !available {
			context.UserError("Name already taken")
			return
		}
	}

	err = deployV2.Deployments().CheckQuota(requestURI.DeploymentID, &opts.QuotaOptions{Update: &requestBody})
	if err != nil {
		var quotaExceededErr sErrors.QuotaExceededError
		if errors.As(err, &quotaExceededErr) {
			context.Forbidden(quotaExceededErr.Error())
			return
		}

		context.ServerError(err, InternalError)
		return
	}

	canUpdate, reason := deployV2.Deployments().CanAddActivity(requestURI.DeploymentID, model.ActivityUpdating)
	if !canUpdate {
		context.Locked(reason)
		return
	}

	jobID := uuid.New().String()
	err = deployV2.Jobs().Create(jobID, auth.User.ID, model.JobUpdateDeployment, version.V2, map[string]interface{}{
		"id":       deployment.ID,
		"params":   requestBody,
		"authInfo": auth,
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

func getDeploymentExternalPort(zoneName string) *int {
	zone := config.Config.GetZone(zoneName)
	if zone == nil {
		return nil
	}

	split := strings.Split(zone.Domains.ParentDeployment, ":")
	if len(split) > 1 {
		port, err := strconv.Atoi(split[1])
		if err != nil {
			return nil
		}

		return &port
	}

	return nil
}
