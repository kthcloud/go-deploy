package v2

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/dto/v2/query"
	"github.com/kthcloud/go-deploy/dto/v2/uri"
	configModels "github.com/kthcloud/go-deploy/models/config"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/models/version"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/sys"
	"github.com/kthcloud/go-deploy/service"
	"github.com/kthcloud/go-deploy/service/clients"
	"github.com/kthcloud/go-deploy/service/core"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	"github.com/kthcloud/go-deploy/service/v2/deployments/opts"
	gpuClaimOpts "github.com/kthcloud/go-deploy/service/v2/gpu_claims/opts"
	teamOpts "github.com/kthcloud/go-deploy/service/v2/teams/opts"
	v12 "github.com/kthcloud/go-deploy/service/v2/utils"
)

// GetDeployment
// @Summary Get deployment
// @Description Get deployment
// @Tags Deployment
// @Produce  json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
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
		context.ServerError(err, ErrAuthInfoNotAvailable)
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
		context.ServerError(err, ErrInternal)
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
// @Produce  json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
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
		context.ServerError(err, ErrAuthInfoNotAvailable)
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
		context.ServerError(err, ErrAuthInfoNotAvailable)
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
// @Security KeycloakOAuth
// @Param body body body.DeploymentCreate true "Deployment body"
// @Success 200 {object} body.DeploymentRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 403 {object} sys.ErrorResponse
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
		context.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	effectiveRole := auth.GetEffectiveRole()
	if effectiveRole == nil {
		context.ServerError(err, ErrInternal)
		return
	}

	deployV2 := service.V2(auth)

	doesNotAlreadyExists, err := deployV2.Deployments().NameAvailable(requestBody.Name)
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	if !doesNotAlreadyExists {
		context.UserError("Deployment already exists")
		return
	}

	for _, gpu := range requestBody.GPUs {
		if strings.TrimSpace(gpu.ClaimName) == "" || strings.TrimSpace(gpu.Name) == "" {
			context.UserError("Invalid gpu claim reference, requires both ClaimName and Name")
			return
		}
	}

	if requestBody.Zone == nil {
		requestBody.Zone = &config.Config.Deployment.DefaultZone
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

		if len(requestBody.GPUs) > 0 {
			if !deployV2.System().ZoneHasCapability(*requestBody.Zone, configModels.ZoneCapabilityDRA) {
				context.Forbidden("Zone does not have dra capability")
				return
			}
		}
	}

	if requestBody.CustomDomain != nil && !effectiveRole.Permissions.UseCustomDomains {
		context.Forbidden("User is not allowed to use custom domains")
		return
	}

	if requestBody.NeverStale && !auth.User.IsAdmin {
		context.Forbidden("User is not allowed to create deployment with neverStale attribute set as true")
		return
	}

	if err := validateGpuRequests(&requestBody.GPUs, *requestBody.Zone, auth, deployV2); err != nil {
		if errors.Is(err, ErrCouldNotGetGpuClaims) {
			context.ServerError(err, ErrCouldNotGetGpuClaims)
			return
		}
		if errors.Is(err, sErrors.NewZoneCapabilityMissingError(*requestBody.Zone, configModels.ZoneCapabilityDRA)) {
			context.Forbidden("Zone lacks DRA capability")
			return
		}
		context.UserError(err.Error())
		return
	}

	err = deployV2.Deployments().CheckQuota("", &opts.QuotaOptions{Create: &requestBody})
	if err != nil {
		var quotaExceededErr sErrors.QuotaExceededError
		if errors.As(err, &quotaExceededErr) {
			context.Forbidden(quotaExceededErr.Error())
			return
		}

		context.ServerError(err, ErrInternal)
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
		context.ServerError(err, ErrInternal)
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
// @Security KeycloakOAuth
// @Param deploymentId path string true "Deployment ID"
// @Success 200 {object} body.DeploymentCreated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 403 {object} sys.ErrorResponse
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
		context.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	deployV2 := service.V2(auth)

	currentDeployment, err := deployV2.Deployments().Get(requestURI.DeploymentID)
	if err != nil {
		context.ServerError(err, ErrInternal)
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

		if errors.Is(err, sErrors.ErrDeploymentNotFound) {
			context.NotFound("Deployment not found")
			return
		}

		context.ServerError(err, ErrInternal)
		return
	}

	jobID := uuid.NewString()
	err = deployV2.Jobs().Create(jobID, auth.User.ID, model.JobDeleteDeployment, version.V2, map[string]interface{}{
		"id":       currentDeployment.ID,
		"ownerId":  auth.User.ID,
		"authInfo": auth,
	})
	if err != nil {
		context.ServerError(err, ErrInternal)
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
// @Security KeycloakOAuth
// @Param deploymentId path string true "Deployment ID"
// @Param body body body.DeploymentUpdate true "Deployment update"
// @Success 200 {object} body.DeploymentUpdated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 403 {object} sys.ErrorResponse
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
		context.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	deployV2 := service.V2(auth)

	deployment, err := deployV2.Deployments().Get(requestURI.DeploymentID, opts.GetOpts{Shared: true})
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}
	if deployment == nil {
		context.NotFound("Deployment not found")
		return
	}

	if requestBody.Name != nil {
		available, err := deployV2.Deployments().NameAvailable(*requestBody.Name)
		if err != nil {
			context.ServerError(err, ErrInternal)
			return
		}

		if !available {
			context.UserError("Name already taken")
			return
		}
	}

	if err := validateGpuRequests(requestBody.GPUs, deployment.Zone, auth, deployV2); err != nil {
		if errors.Is(err, ErrCouldNotGetGpuClaims) {
			context.ServerError(err, ErrCouldNotGetGpuClaims)
			return
		}
		if errors.Is(err, sErrors.NewZoneCapabilityMissingError(deployment.Zone, configModels.ZoneCapabilityDRA)) {
			context.Forbidden("Zone lacks DRA capability")
			return
		}
		context.UserError(err.Error())
		return
	}

	if requestBody.NeverStale != nil && !auth.User.IsAdmin {
		context.Forbidden("User is not allowed to modify the neverStale value")
		return
	}

	err = deployV2.Deployments().CheckQuota(requestURI.DeploymentID, &opts.QuotaOptions{Update: &requestBody})
	if err != nil {
		var quotaExceededErr sErrors.QuotaExceededError
		if errors.As(err, &quotaExceededErr) {
			context.Forbidden(quotaExceededErr.Error())
			return
		}

		context.ServerError(err, ErrInternal)
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
		context.ServerError(err, ErrInternal)
		return
	}

	context.Ok(body.DeploymentUpdated{
		ID:    deployment.ID,
		JobID: &jobID,
	})
}

func validateGpuRequests(gpus *[]body.DeploymentGPU, zone string, auth *core.AuthInfo, deployV2 clients.V2) error {
	if gpus != nil {
		if len(*gpus) > 0 {

			if !deployV2.System().ZoneHasCapability(zone, configModels.ZoneCapabilityDRA) {
				return sErrors.NewZoneCapabilityMissingError(zone, configModels.ZoneCapabilityDRA)
			}

			roles := make([]string, 0, 2)
			if role := auth.GetEffectiveRole(); role != nil {
				roles = append(roles, role.Name)
			}
			if auth.User != nil && auth.User.IsAdmin {
				roles = append(roles, "admin")
			}
			claimReqMap := make(map[string][]string, len(*gpus))
			for _, gpu := range *gpus {
				if reqs, found := claimReqMap[gpu.ClaimName]; found {
					claimReqMap[gpu.ClaimName] = append(reqs, gpu.Name)
				} else {
					claimReqMap[gpu.ClaimName] = []string{gpu.Name}
				}
			}
			names := slices.AppendSeq(make([]string, 0, len(claimReqMap)), maps.Keys(claimReqMap))
			claims, err := deployV2.GpuClaims().List(gpuClaimOpts.ListOpts{
				Names: &names,
				Zone:  &zone,
				Roles: &roles,
			})
			if err != nil {
				return errors.Join(err, ErrCouldNotGetGpuClaims)
			}
			availableClaimReqMap := make(map[string][]string, len(claims))
			for _, claim := range claims {
				if reqs, found := availableClaimReqMap[claim.Name]; found {
					availableClaimReqMap[claim.Name] = append(reqs, slices.AppendSeq(make([]string, 0, len(claim.Requested)), maps.Keys(claim.Requested))...)
				} else {
					availableClaimReqMap[claim.Name] = slices.AppendSeq(make([]string, 0, len(claim.Requested)), maps.Keys(claim.Requested))
				}
			}
			missingClaims := make([]string, 0)
			missingRequests := make([]string, 0)
			for claimName, reqsName := range claimReqMap {
				if availableReqs, found := availableClaimReqMap[claimName]; found {
					for _, req := range reqsName {
						if !slices.Contains(availableReqs, req) {
							missingRequests = append(missingRequests,
								fmt.Sprintf("%s:%s", claimName, req),
							)
						}
					}
				} else {
					missingClaims = append(missingClaims, claimName)
				}
			}
			var errs []error

			if len(missingClaims) > 0 {
				for _, c := range missingClaims {
					errs = append(errs,
						fmt.Errorf("missing GPU claim: %s", c),
					)
				}
			}

			if len(missingRequests) > 0 {
				for _, r := range missingRequests {
					errs = append(errs,
						fmt.Errorf("missing GPU request in claim: %s", r),
					)
				}
			}

			if len(errs) > 0 {
				return errors.Join(errs...)
			}
		}
	}
	return nil
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
