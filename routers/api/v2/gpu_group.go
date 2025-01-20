package v2

import (
	"github.com/gin-gonic/gin"
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/dto/v2/query"
	"github.com/kthcloud/go-deploy/dto/v2/uri"
	"github.com/kthcloud/go-deploy/pkg/sys"
	"github.com/kthcloud/go-deploy/service"
	"github.com/kthcloud/go-deploy/service/v2/utils"
	"github.com/kthcloud/go-deploy/service/v2/vms/opts"
)

// GetGpuGroup
// @Summary Get GPU group
// @Description Get GPU group
// @Tags GpuGroup
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param gpuGroupId path string true "GPU group ID"
// @Success 200 {object} body.GpuGroupRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/gpuGroups/{gpuGroupId} [get]
func GetGpuGroup(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.GpuGroupGet
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

	gpuGroup, err := deployV2.VMs().GpuGroups().Get(requestURI.GpuGroupID)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if gpuGroup == nil {
		context.NotFound("GPU group not found")
		return
	}

	leases, err := deployV2.VMs().GpuLeases().Count(opts.ListGpuLeaseOpts{
		GpuGroupID: &gpuGroup.ID,
	})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	context.Ok(gpuGroup.ToDTO(leases))
}

// ListGpuGroups
// @Summary List GPU groups
// @Description List GPU groups
// @Tags GpuGroup
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {array} body.GpuGroupRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/gpuGroups [get]
func ListGpuGroups(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.GpuGroupList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var requestURI uri.GpuGroupList
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

	gpuGroups, err := deployV2.VMs().GpuGroups().List(opts.ListGpuGroupOpts{
		Pagination: utils.GetOrDefaultPagination(requestQuery.Pagination),
	})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if gpuGroups == nil {
		context.Ok([]interface{}{})
		return
	}

	dtoGpuGroups := make([]body.GpuGroupRead, len(gpuGroups))
	for i, gpuGroup := range gpuGroups {
		leases, err := deployV2.VMs().GpuLeases().Count(opts.ListGpuLeaseOpts{
			GpuGroupID: &gpuGroup.ID,
		})
		if err != nil {
			context.ServerError(err, InternalError)
			return
		}

		dtoGpuGroups[i] = gpuGroup.ToDTO(leases)
	}

	context.Ok(dtoGpuGroups)
}
