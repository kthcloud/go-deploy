package v2

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/dto/v2/query"
	"github.com/kthcloud/go-deploy/dto/v2/uri"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/models/version"
	"github.com/kthcloud/go-deploy/pkg/sys"
	"github.com/kthcloud/go-deploy/service"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	"github.com/kthcloud/go-deploy/service/v2/utils"
	"github.com/kthcloud/go-deploy/service/v2/vms/opts"
)

// GetGpuLease
// @Summary Get GPU lease
// @Description Get GPU lease
// @Tags GpuLease
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param gpuLeaseId path string true "GPU lease ID"
// @Success 200 {object} body.GpuLeaseRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/gpuLeases/{gpuLeaseId} [get]
func GetGpuLease(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.GpuLeaseGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	gpuLease, err := service.V2(auth).VMs().GpuLeases().Get(requestURI.GpuLeaseID)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if gpuLease == nil {
		context.NotFound("GPU lease not found")
		return
	}

	position, err := service.V2(auth).VMs().GpuLeases().GetQueuePosition(gpuLease.ID)
	if err != nil {
		if errors.Is(err, sErrors.GpuLeaseNotFoundErr) {
			position = -1
		}

		context.ServerError(err, InternalError)
	}

	context.Ok(gpuLease.ToDTO(position))
}

// ListGpuLeases
// @Summary List GPU leases
// @Description List GPU leases
// @Tags GpuLease
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param all query bool false "List all"
// @Param vmId query string false "Filter by VM ID"
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {array} body.GpuLeaseRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/gpuLeases [get]
func ListGpuLeases(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.GpuLeaseList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var requestURI uri.GpuLeaseList
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	var userID *string
	if !requestQuery.All {
		userID = &auth.User.ID
	}

	gpuLeases, err := service.V2(auth).VMs().GpuLeases().List(opts.ListGpuLeaseOpts{
		Pagination: utils.GetOrDefaultPagination(requestQuery.Pagination),
		UserID:     userID,
	})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if gpuLeases == nil {
		context.Ok([]interface{}{})
		return
	}

	dtoGpuLeases := make([]body.GpuLeaseRead, 0)
	for _, gpuLease := range gpuLeases {
		queuePosition, err := service.V2(auth).VMs().GpuLeases().GetQueuePosition(gpuLease.ID)
		if err != nil {
			switch {
			case errors.Is(err, sErrors.GpuLeaseNotFoundErr):
				continue
			case errors.Is(err, sErrors.GpuGroupNotFoundErr):
				continue
			}

			queuePosition = -1
		}

		dtoGpuLeases = append(dtoGpuLeases, gpuLease.ToDTO(queuePosition))
	}

	context.Ok(dtoGpuLeases)
}

// CreateGpuLease
// @Summary Create GPU Lease
// @Description Create GPU lease
// @Tags GpuLease
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param body body body.GpuLeaseCreate true "GPU lease"
// @Success 200 {object} body.GpuLeaseCreated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/gpuLeases [post]
func CreateGpuLease(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.GpuLeaseCreate
	if err := context.GinContext.ShouldBindQuery(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var requestBody body.GpuLeaseCreate
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

	allowedToLease := auth.GetEffectiveRole().Permissions.UseGPUs
	if !allowedToLease {
		context.UserError("User not allowed to lease GPUs")
		return
	}

	groupExists, err := deployV2.VMs().GpuGroups().Exists(requestBody.GpuGroupID)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if !groupExists {
		context.NotFound("GPU group not found")
		return
	}

	// Right now we only allow a single lease per user, this can be updated in the future
	anyGpuLease, err := deployV2.VMs().GpuLeases().Count(opts.ListGpuLeaseOpts{
		UserID: &auth.User.ID,
	})

	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if anyGpuLease > 0 {
		context.UserError("User already has a GPU lease")
		return
	}

	gpuLeaseID := uuid.New().String()
	jobID := uuid.New().String()
	err = deployV2.Jobs().Create(jobID, auth.User.ID, model.JobCreateGpuLease, version.V2, map[string]interface{}{
		"id":       gpuLeaseID,
		"userId":   auth.User.ID,
		"params":   requestBody,
		"authInfo": auth,
	})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	context.Ok(body.GpuLeaseCreated{
		ID:    gpuLeaseID,
		JobID: jobID,
	})
}

// UpdateGpuLease
// @Summary Update GPU lease
// @Description Update GPU lease
// @Tags GpuLease
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param gpuLeaseId path string true "GPU lease ID"
// @Param body body body.GpuLeaseUpdate true "GPU lease"
// @Success 200 {object} body.GpuLeaseUpdated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/gpuLeases/{gpuLeaseId} [post]
func UpdateGpuLease(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.GpuLeaseUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var requestBody body.GpuLeaseUpdate
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

	gpuLease, err := deployV2.VMs().GpuLeases().Get(requestURI.GpuLeaseID)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if gpuLease == nil {
		context.NotFound("GPU lease not found")
		return
	}

	// If the update includes activating the lease, we ensure it is allowed
	if requestBody.VmID != nil && gpuLease.AssignedAt == nil {
		context.UserError("GPU lease is not assigned")
		return
	}

	jobID := uuid.New().String()
	err = deployV2.Jobs().Create(jobID, auth.User.ID, model.JobUpdateGpuLease, version.V2, map[string]interface{}{
		"id":       gpuLease.ID,
		"params":   requestBody,
		"authInfo": auth,
	})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	context.Ok(body.GpuLeaseUpdated{
		ID:    gpuLease.ID,
		JobID: jobID,
	})
}

// DeleteGpuLease
// @Summary Delete GPU lease
// @Description Delete GPU lease
// @Tags GpuLease
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param gpuLeaseId path string true "GPU lease ID"
// @Success 200 {object} body.GpuLeaseDeleted
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/gpuLeases/{gpuLeaseId} [delete]
func DeleteGpuLease(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.GpuLeaseDelete
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

	gpuLease, err := deployV2.VMs().GpuLeases().Get(requestURI.GpuLeaseID)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if gpuLease == nil {
		context.NotFound("GPU lease not found")
		return
	}

	jobID := uuid.New().String()
	err = deployV2.Jobs().Create(jobID, auth.User.ID, model.JobDeleteGpuLease, version.V2, map[string]interface{}{
		"id":       gpuLease.ID,
		"authInfo": auth,
	})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	context.Ok(body.GpuLeaseDeleted{
		ID:    gpuLease.ID,
		JobID: jobID,
	})
}
