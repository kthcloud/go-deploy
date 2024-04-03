package v2

import (
	"github.com/gin-gonic/gin"
	"go-deploy/dto/v2/body"
	"go-deploy/dto/v2/query"
	"go-deploy/dto/v2/uri"
	"go-deploy/pkg/sys"
	"go-deploy/routers/api/v1"
	"go-deploy/service"
	"go-deploy/service/v2/utils"
	"go-deploy/service/v2/vms/opts"
	"math/rand"
)

// GetGpuGroup
// @Summary Get GPU group
// @Description Get GPU group
// @Tags VM
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param gpuGroupId path string true "GPU group ID"
// @Success 200 {object} body.GpuGroupRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /gpuGroups/{gpuGroupId} [get]
func GetGpuGroup(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.GpuGroupGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	gpuGroup, err := service.V2(auth).VMs().GpuGroups().Get(requestURI.GpuGroupID)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if gpuGroup == nil {
		context.NotFound("GPU group not found")
		return
	}

	// TODO: Replace with proper check how many leases
	// Temporary to test API
	leases := rand.Int() % gpuGroup.Total

	context.Ok(gpuGroup.ToDTO(leases))
}

// ListGpuGroups
// @Summary List GPU groups
// @Description List GPU groups
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
// @Router /gpuGroups [get]
func ListGpuGroups(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.GpuGroupList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	var requestURI uri.GpuGroupList
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	gpuGroups, err := service.V2(auth).VMs().GpuGroups().List(opts.ListGpuGroupOpts{
		Pagination: utils.GetOrDefaultPagination(requestQuery.Pagination),
	})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if gpuGroups == nil {
		context.Ok([]interface{}{})
		return
	}

	dtoGpuGroups := make([]body.GpuGroupRead, len(gpuGroups))
	for i, group := range gpuGroups {

		// TODO: Replace with proper check how many leases
		// Temporary to test API
		leases := rand.Int() % group.Total

		dtoGpuGroups[i] = group.ToDTO(leases)
	}

	context.Ok(dtoGpuGroups)
}
