package v1_vm

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/vm_service"
	"net/http"
)

// GetGpuList
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
func GetGpuList(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.GpuList
	if err := context.GinContext.Bind(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	gpus, err := vm_service.GetAllGPUs(requestQuery.OnlyShowAvailable, auth.GetEffectiveRole().Permissions.UsePrivilegedGPUs)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	dtoGPUs := make([]body.GpuRead, len(gpus))
	for i, gpu := range gpus {
		dtoGPUs[i] = gpu.ToDTO(false)
	}

	context.JSONResponse(200, dtoGPUs)
}
