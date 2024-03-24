package v2

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/dto/v2/body"
	"go-deploy/dto/v2/query"
	"go-deploy/dto/v2/uri"
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	"go-deploy/service/v2/utils"
	"go-deploy/service/v2/vms/opts"
)

// GetGpuLease
// @Summary GetVM GPU lease
// @Description GetVM GPU lease
// @Tags VM
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param gpuLeaseId path string true "GPU lease ID"
// @Success 200 {object} body.GpuLeaseRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /gpuLeases/{gpuLeaseId} [get]
func GetGpuLease(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.GpuLeaseGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	gpuLease, err := service.V2(auth).VMs().GpuLeases().Get(requestURI.GpuLeaseID)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if gpuLease == nil {
		context.NotFound("GPU lease not found")
		return
	}

	context.Ok(gpuLease.ToDTO())
}

// ListGpuLeases
// @Summary GetVM GPU lease list
// @Description GetVM GPU lease list
// @Tags VM
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param vmId query string false "VM ID"
// @Success 200 {array} body.
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /gpuLeases [get]
func ListGpuLeases(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.GpuLeaseList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	var requestURI uri.GpuLeaseList
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	gpuLeases, err := service.V2(auth).VMs().GpuLeases().List(opts.ListGpuLeaseOpts{
		Pagination: utils.GetOrDefaultPagination(requestQuery.Pagination),
	})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if gpuLeases == nil {
		context.Ok([]interface{}{})
		return
	}

	dtoGpuLeases := make([]body.GpuLeaseRead, len(gpuLeases))
	for i, gpuLease := range gpuLeases {
		dtoGpuLeases[i] = gpuLease.ToDTO()
	}

	context.Ok(dtoGpuLeases)
}

// CreateGpuLease
// @Summary CreateVM GPU lease
// @Description CreateVM GPU lease
// @Tags VM
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body body.GpuLeaseCreate true "GPU lease"
// @Success 200 {object} body.GpuLeaseRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /gpuLeases [post]
func CreateGpuLease(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.GpuLeaseCreate
	if err := context.GinContext.ShouldBindQuery(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	var requestBody body.GpuLeaseCreate
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
	deployV2 := service.V2(auth)

	canAccess, err := deployV2.VMs().IsAccessible(requestQuery.VmID)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if !canAccess {
		context.Forbidden("User does not have access to the VM")
		return
	}

	// TODO: Check if GPU type is valid

	gpuLeaseID := uuid.New().String()
	jobID := uuid.New().String()
	err = deployV1.Jobs().Create(jobID, auth.UserID, model.JobCreateGpuLease, version.V2, map[string]interface{}{
		"id":       gpuLeaseID,
		"vmId":     requestQuery.VmID,
		"userId":   auth.UserID,
		"authInfo": *auth,
		"params":   requestBody,
	})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(body.GpuLeaseCreated{
		ID:    gpuLeaseID,
		JobID: jobID,
	})
}

// DeleteGpuLease
// @Summary DeleteVM GPU lease
// @Description DeleteVM GPU lease
// @Tags VM
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param gpuLeaseId path string true "GPU lease ID"
// @Success 200 {object} body.GpuLeaseRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /gpuLeases/{gpuLeaseId} [delete]
func DeleteGpuLease(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.GpuLeaseDelete
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
	deployV2 := service.V2(auth)

	gpuLease, err := deployV2.VMs().GpuLeases().Get(requestURI.GpuLeaseID)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if gpuLease == nil {
		context.NotFound("GPU lease not found")
		return
	}

	jobID := uuid.New().String()
	err = deployV1.Jobs().Create(jobID, auth.UserID, model.JobDeleteGpuLease, version.V2, map[string]interface{}{
		"id": gpuLease.ID,
	})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(gpuLease.ToDTO())
}
