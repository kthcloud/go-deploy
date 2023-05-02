package v1_vm

import (
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	"go-deploy/models/vm"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/vm_service"
	"net/http"
)

func GetGpuList(c *gin.Context) {
	context := app.NewContext(c)

	var requestQuery query.GpuList
	if err := context.GinContext.Bind(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(&requestQuery, err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	gpus, err := vm_service.GetAllGPUs(requestQuery.OnlyShowAvailable, auth.IsAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	dtoGPUs := make([]body.GpuRead, len(gpus))
	for i, gpu := range gpus {
		dtoGPUs[i] = gpu.ToDto()
	}

	context.JSONResponse(200, dtoGPUs)
}

func AttachGPU(c *gin.Context) {
	context := app.NewContext(c)

	var requestURI uri.GpuAttach
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(&requestURI, err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	gpuID, err := getGpuID(&requestURI)
	if err != nil {
		context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceValidationFailed, "Invalid GPU ID")
		return
	}

	current, err := vm_service.GetByID(auth.UserID, requestURI.VmID, auth.IsAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if current == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("VM with ID %s not found", requestURI.VmID))
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
		gpu, err = vm_service.GetAnyAvailableGPU(auth.IsPowerUser)
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
			return
		}

		if gpu == nil {
			context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotAvailable, "No available GPUs")
			return
		}
	} else {
		gpu, err = vm_service.GetGpuByID(gpuID, auth.IsPowerUser)
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

	vm_service.AttachGPU(gpu.ID, current.ID, auth.UserID)

	// the returned gpu might not actually get attached, but it will work in most cases
	context.JSONResponse(http.StatusCreated, gpu.ToDto())
}

func DetachGPU(c *gin.Context) {
	context := app.NewContext(c)

	var requestURI uri.GpuDetach
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(&requestURI, err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	current, err := vm_service.GetByID(auth.UserID, requestURI.VmID, auth.IsAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, fmt.Sprintf("Failed to get VM: %s", err))
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

	vm_service.DetachGPU(current.ID, auth.UserID)

	context.OkDeleted()
}

func decodeGpuID(gpuID string) (string, error) {
	res, err := base64.StdEncoding.DecodeString(gpuID)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func getGpuID(requestURI *uri.GpuAttach) (string, error) {
	if requestURI.GpuID == "" {
		return "any", nil
	}

	return decodeGpuID(requestURI.GpuID)
}
