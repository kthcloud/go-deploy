package v1_vm

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto"
	"go-deploy/models/vm"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/validator"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/vm_service"
	"net/http"
)

func GetGpuList(c *gin.Context) {
	context := app.NewContext(c)

	rules := validator.MapData{
		"available": []string{"bool"},
	}

	validationErrors := context.ValidateQueryParams(&rules)
	if len(validationErrors) > 0 {
		context.ResponseValidationError(validationErrors)
		return
	}

	onlyAvailable := context.GinContext.Query("available") == "true"
	isGpuUser := v1.IsGpuUser(&context)

	var gpus []vm.GPU
	var err error

	gpus, err = vm_service.GetAllGPUs(onlyAvailable, isGpuUser)

	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	dtoGPUs := make([]dto.GpuRead, len(gpus))
	for i, gpu := range gpus {
		dtoGPUs[i] = gpu.ToDto()
	}

	context.JSONResponse(200, dtoGPUs)
}
