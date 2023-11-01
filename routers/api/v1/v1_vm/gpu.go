package v1_vm

import (
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/vm_service"
)

// ListGPUs
// @Summary Get list of GPUs
// @Description Get list of GPUs
// @Tags VM
// @Accept  json
// @Produce  json
// @Param available query bool false "Only show available GPUs"
// @Success 200 {array} body.GpuRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /vms/gpus [get]
func ListGPUs(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.GpuList
	if err := context.GinContext.Bind(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	gpus, err := vm_service.ListGPUs(requestQuery.OnlyShowAvailable, auth.GetEffectiveRole().Permissions.UsePrivilegedGPUs)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	dtoGPUs := make([]body.GpuRead, len(gpus))
	for i, gpu := range gpus {
		dtoGPUs[i] = gpu.ToDTO(false)
	}

	context.Ok(dtoGPUs)
}
