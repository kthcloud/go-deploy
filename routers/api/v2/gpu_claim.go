package v2

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/dto/v2/query"
	"github.com/kthcloud/go-deploy/dto/v2/uri"
	configModels "github.com/kthcloud/go-deploy/models/config"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/models/version"
	"github.com/kthcloud/go-deploy/pkg/sys"
	"github.com/kthcloud/go-deploy/service"
	"github.com/kthcloud/go-deploy/service/v2/gpu_claims/opts"
	"github.com/kthcloud/go-deploy/service/v2/utils"
)

// GetGpuClaim
// @Summary Get GPU claim
// @Description Get GPU claim
// @Tags GpuClaim
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param gpuClaimId path string true "GPU claim ID"
// @Success 200 {object} body.GpuClaimRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/gpuClaims/{gpuClaimId} [get]
func GetGpuClaim(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.GpuClaimGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	if auth.User == nil || !auth.User.IsAdmin {
		context.Forbidden("User is not allowed to get GpuClaim")
		return
	}

	deployV2 := service.V2(auth)

	GpuClaim, err := deployV2.GpuClaims().Get(requestURI.GpuClaimID)
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	if GpuClaim == nil {
		context.NotFound("GPU claim not found")
		return
	}

	context.Ok(GpuClaim.ToDTO())
}

// ListGpuClaims
// @Summary List GPU claims
// @Description List GPU claims
// @Tags GpuClaim
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Param detailed query bool false "Admin detailed list"
// @Success 200 {array} body.GpuClaimRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/gpuClaims [get]
func ListGpuClaims(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.GpuClaimList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var requestURI uri.GpuClaimList
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	if requestQuery.Detailed && (auth.User == nil || !auth.User.IsAdmin) {
		context.ErrorResponse(http.StatusForbidden, 403, "only admins can access detailed view of gpu claims")
	}

	roles := make([]string, 0, 2)
	if role := auth.GetEffectiveRole(); role != nil {
		roles = append(roles, role.Name)
	}
	if auth.User.IsAdmin {
		roles = append(roles, "admin")
	}

	deployV2 := service.V2(auth)

	GpuClaims, err := deployV2.GpuClaims().List(opts.ListOpts{
		Pagination: utils.GetOrDefaultPagination(requestQuery.Pagination),
		Roles:      &roles,
	})
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	if GpuClaims == nil {
		context.Ok([]any{})
		return
	}

	dtoGpuClaims := make([]body.GpuClaimRead, len(GpuClaims))
	for i, GpuClaim := range GpuClaims {
		if requestQuery.Detailed {
			dtoGpuClaims[i] = GpuClaim.ToDTO()
		} else {
			dtoGpuClaims[i] = GpuClaim.ToBriefDTO()
		}
	}

	context.Ok(dtoGpuClaims)
}

// CreateGpuClaim
// @Summary Create GpuClaim
// @Description Create GpuClaim
// @Tags GpuClaim
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param body body body.GpuClaimCreate true "GpuClaim body"
// @Success 200 {object} body.GpuClaimCreated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 403 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/gpuClaims [post]
func CreateGpuClaim(c *gin.Context) {
	context := sys.NewContext(c)

	var requestBody body.GpuClaimCreate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	if auth.User == nil || !auth.User.IsAdmin {
		context.Forbidden("User is not allowed to create GpuClaim")
		return
	}

	deployV2 := service.V2(auth)

	/*doesNotAlreadyExists, err := deployV2.GpuClaims().NameAvailable(requestBody.Name)
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	if !doesNotAlreadyExists {
		context.UserError("GpuClaim already exists")
		return
	}*/

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

		if !deployV2.System().ZoneHasCapability(*requestBody.Zone, configModels.ZoneCapabilityDRA) {
			context.Forbidden("Zone does not have dra capability")
			return
		}
	}

	GpuClaimID := uuid.New().String()
	jobID := uuid.New().String()
	err = deployV2.Jobs().Create(jobID, auth.User.ID, model.JobCreateGpuClaim, version.V2, map[string]any{
		"id":       GpuClaimID,
		"params":   requestBody,
		"authInfo": auth,
	})
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	context.Ok(body.GpuClaimCreated{
		ID:    GpuClaimID,
		JobID: jobID,
	})
}

// DeleteGpuClaim
// @Summary Delete GpuClaim
// @Description Delete GpuClaim
// @Tags GpuClaim
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param gpuClaimId path string true "GpuClaim ID"
// @Success 200 {object} body.GpuClaimCreated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 403 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/gpuClaims/{gpuClaimId} [delete]
func DeleteGpuClaim(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.GpuClaimDelete
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	if auth.User == nil || !auth.User.IsAdmin {
		context.Forbidden("GpuClaims can only be deleted by admins")
		return
	}

	deployV2 := service.V2(auth)

	currentGpuClaim, err := deployV2.GpuClaims().Get(requestURI.GpuClaimID)
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	if currentGpuClaim == nil {
		context.NotFound("GpuClaim not found")
		return
	}

	/*err = deployV2.GpuClaims().StartActivity(requestURI.GpuClaimID, model.ActivityBeingDeleted)
	if err != nil {
		var failedToStartActivityErr sErrors.FailedToStartActivityError
		if errors.As(err, &failedToStartActivityErr) {
			context.Locked(failedToStartActivityErr.Error())
			return
		}

		if errors.Is(err, sErrors.ErrResourceNotFound) {
			context.NotFound("GpuClaim not found")
			return
		}

		context.ServerError(err, ErrInternal)
		return
	}*/

	jobID := uuid.NewString()
	err = deployV2.Jobs().Create(jobID, auth.User.ID, model.JobDeleteGpuClaim, version.V2, map[string]any{
		"id":       currentGpuClaim.ID,
		"authInfo": auth,
	})
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	context.Ok(body.GpuClaimCreated{
		ID:    currentGpuClaim.ID,
		JobID: jobID,
	})
}
