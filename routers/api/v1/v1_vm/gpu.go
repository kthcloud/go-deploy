package v1_vm

import (
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	jobModel "go-deploy/models/sys/job"
	vmModel "go-deploy/models/sys/vm"
	gpuModel "go-deploy/models/sys/vm/gpu"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/job_service"
	"go-deploy/service/vm_service"
	"net/http"
)

// GetGpuList
// @Summary Get list of GPUs
// @Description Get list of GPUs
// @Tags VM
// @Accept  json
// @Produce  json
// @Param onlyShowAvailable query bool false "Only show available GPUs"
// @Success 200 {array} body.GpuRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
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

	gpus, err := vm_service.GetAllGPUs(requestQuery.OnlyShowAvailable, auth.IsAdmin())
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	dtoGPUs := make([]body.GpuRead, len(gpus))
	for i, gpu := range gpus {
		dtoGPUs[i] = gpu.ToDto(false)
	}

	context.JSONResponse(200, dtoGPUs)
}

// AttachGPU
// @Summary Attach GPU to VM
// @Description Attach GPU to VM
// @Tags VM
// @Accept  json
// @Produce  json
// @Param vmId path string true "VM ID"
// @Param gpuId path string false "GPU ID"
// @Success 200 {object} body.VmRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /api/v1/vms/{vmId}/attachGpu/{gpuId} [post]
func AttachGPU(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.GpuAttach
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
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

	vm, err := vm_service.GetByID(auth.UserID, requestURI.VmID, auth.IsAdmin())
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if vm == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("VM with ID %s not found", requestURI.VmID))
		return
	}

	// if a request for "any" comes in while already attached to a gpu, assume it's a request to reattach
	if gpuID == "any" && vm.GpuID != "" {
		gpuID = vm.GpuID
	}

	if vm.GpuID != "" && vm.GpuID != gpuID {
		context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceNotCreated, "Resource already has a GPU attached")
		return
	}

	var gpus []gpuModel.GPU
	if gpuID == "any" {
		gpus, err = vm_service.GetAllAvailableGPU(auth.IsPowerUser())
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get available GPU: %s", err))
			return
		}

		if gpus == nil {
			context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotAvailable, "No available GPUs")
			return
		}
	} else {
		gpu, err := vm_service.GetGpuByID(gpuID, auth.IsPowerUser())
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
				context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to check if GPU is available: %s", err))
				return
			}

			if !available {
				context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotAvailable, "GPU not available")
				return
			}
		}

		gpus = []gpuModel.GPU{*gpu}
	}

	if len(gpus) == 0 {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotAvailable, "No available GPUs")
		return
	}

	// do this check to give a nice error to the user if the gpu cannot be attached
	// otherwise it will be silently ignored
	if len(gpus) == 1 {
		canStartOnHost, reason, err := vm_service.CanStartOnHost(vm.ID, gpus[0].Host)
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to check if VM can start on host: %s", err))
			return
		}

		if !canStartOnHost {
			if reason == "" {
				reason = "VM could not on the host"
			}

			context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceNotUpdated, reason)
			return
		}
	}

	gpuIds := make([]string, len(gpus))
	for i, gpu := range gpus {
		gpuIds[i] = gpu.ID
	}

	started, reason, err := vm_service.StartActivity(vm.ID, vmModel.ActivityAttachingGPU)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to start activity: %s", err))
		return
	}

	if !started {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceNotReady, reason)
		return
	}

	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeAttachGpuToVM, map[string]interface{}{
		"id":     vm.ID,
		"gpuIds": gpuIds,
		"userId": auth.UserID,
	})

	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to create job: %s", err))
		return
	}

	context.JSONResponse(http.StatusOK, body.GpuAttached{
		ID:    vm.ID,
		JobID: jobID,
	})
}

// DetachGPU
// @Summary Detach GPU from VM
// @Description Detach GPU from VM
// @Tags VM
// @Accept  json
// @Produce  json
// @Param vmId path string true "VM ID"
// @Success 200 {object} body.VmRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /api/v1/vms/{vmId}/detachGpu [post]
func DetachGPU(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.GpuDetach
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	current, err := vm_service.GetByID(auth.UserID, requestURI.VmID, auth.IsAdmin())
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, fmt.Sprintf("Failed to get VM: %s", err))
		return
	}

	if current == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, "Resource not found")
		return
	}

	if current.GpuID == "" {
		context.ErrorResponse(http.StatusNotModified, status_codes.ResourceNotUpdated, "Resource does not have a GPU attached")
		return
	}

	started, reason, err := vm_service.StartActivity(current.ID, vmModel.ActivityDetachingGPU)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to start activity: %s", err))
		return
	}

	if !started {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceNotReady, reason)
		return
	}

	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeDetachGpuFromVM, map[string]interface{}{
		"id":     current.ID,
		"userId": auth.UserID,
	})
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to create job: %s", err))
		return
	}

	context.JSONResponse(http.StatusOK, body.GpuDetached{
		ID:    current.ID,
		JobID: jobID,
	})
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
