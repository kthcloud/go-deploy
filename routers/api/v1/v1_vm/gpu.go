package v1_vm

import (
	"encoding/base64"
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

func AttachGPU(c *gin.Context) {
	context := app.NewContext(c)

	rules := validator.MapData{
		"vmId": []string{
			"required",
			"uuid_v4",
		},
		"gpuId": []string{},
	}

	validationErrors := context.ValidateParams(&rules)
	if len(validationErrors) > 0 {
		context.ResponseValidationError(validationErrors)
		return
	}

	token, err := context.GetKeycloakToken()
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}
	userID := token.Sub
	vmID := context.GinContext.Param("vmId")
	gpuID, err := getGpuID(&context)
	if err != nil {
		context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceValidationFailed, "Invalid GPU ID")
		return
	}

	isAdmin := v1.IsAdmin(&context)
	isGpuUser := v1.IsGpuUser(&context)

	current, err := vm_service.GetByID(userID, vmID, isAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if current == nil {
		context.NotFound()
		return
	}

	if current.BeingCreated {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingCreated, "Resource is currently being created")
		return
	}

	if current.BeingDeleted {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingDeleted, "Resource is currently being deleted")
		return
	}

	// if a request for "any" comes in while already attached to a gpu, assume it's a request to reattach
	if gpuID == "any" && current.GpuID != "" {
		gpuID = current.GpuID
	}

	if current.GpuID != "" && current.GpuID != gpuID {
		context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceNotCreated, "Resource already has a GPU attached")
		return
	}

	var gpu *vm.GPU
	if gpuID == "any" {
		gpu, err = vm_service.GetAnyAvailableGPU(isGpuUser)
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
			return
		}

		if gpu == nil {
			context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotAvailable, "No available GPUs")
			return
		}
	} else {
		gpu, err = vm_service.GetGpuByID(gpuID, isGpuUser)
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
			return
		}

		if gpu == nil {
			context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, "GPU not found")
			return
		}

		if gpu.Lease.VmID == "" {
			// we still need to check if the gpu is available since the database is not guaranteed to know
			available, err := vm_service.IsGpuAvailable(gpu)
			if err != nil {
				context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
				return
			}

			if !available {
				context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotAvailable, "GPU not available")
				return
			}
		}

	}

	vm_service.AttachGPU(gpu.ID, current.ID, userID)

	// the returned gpu might not actually get attached, but it will work in most cases
	context.JSONResponse(http.StatusCreated, gpu.ToDto())
}

func DetachGPU(c *gin.Context) {
	context := app.NewContext(c)

	rules := validator.MapData{
		"vmId": []string{
			"required",
			"uuid_v4",
		},
	}

	validationErrors := context.ValidateParams(&rules)
	if len(validationErrors) > 0 {
		context.ResponseValidationError(validationErrors)
		return
	}

	token, err := context.GetKeycloakToken()
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}
	userID := token.Sub
	vmID := context.GinContext.Param("vmId")
	if err != nil {
		context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceValidationFailed, "Invalid GPU ID")
		return
	}

	isAdmin := v1.IsAdmin(&context)

	current, err := vm_service.GetByID(userID, vmID, isAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if current == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, "Resource not found")
		return
	}

	if current.BeingCreated {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingCreated, "Resource is currently being created")
		return
	}

	if current.BeingDeleted {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingDeleted, "Resource is currently being deleted")
		return
	}

	if current.GpuID == "" {
		context.ErrorResponse(http.StatusNotModified, status_codes.ResourceNotUpdated, "Resource does not have a GPU attached")
		return
	}

	vm_service.DetachGPU(current.ID, userID)

	context.OkDeleted()
}

func decodeGpuID(gpuID string) (string, error) {
	res, err := base64.StdEncoding.DecodeString(gpuID)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func getGpuID(context *app.ClientContext) (string, error) {
	gpuID := context.GinContext.Param("gpuId")
	if gpuID == "" {
		return "any", nil
	}

	return decodeGpuID(gpuID)
}
