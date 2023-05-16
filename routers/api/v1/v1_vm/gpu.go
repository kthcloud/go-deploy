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
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
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
// @Failure 400 {object} app.ErrorResponse
// @Failure 404 {object} app.ErrorResponse
// @Failure 423 {object} app.ErrorResponse
// @Failure 500 {object} app.ErrorResponse
func GetGpuList(c *gin.Context) {
	context := app.NewContext(c)

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
		dtoGPUs[i] = gpu.ToDto()
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
// @Param gpuId path string true "GPU ID"
// @Success 200 {object} body.VmRead
// @Failure 400 {object} app.ErrorResponse
// @Failure 404 {object} app.ErrorResponse
// @Failure 423 {object} app.ErrorResponse
// @Failure 500 {object} app.ErrorResponse
// @Router /api/v1/vms/{vmId}/gpu [post]
func AttachGPU(c *gin.Context) {
	context := app.NewContext(c)

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

	current, err := vm_service.GetByID(auth.UserID, requestURI.VmID, auth.IsAdmin())
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if current == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("VM with ID %s not found", requestURI.VmID))
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

	var gpu *gpuModel.GPU
	if gpuID == "any" {
		gpu, err = vm_service.GetAnyAvailableGPU(auth.IsPowerUser())
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get available GPU: %s", err))
			return
		}

		if gpu == nil {
			context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotAvailable, "No available GPUs")
			return
		}
	} else {
		gpu, err = vm_service.GetGpuByID(gpuID, auth.IsPowerUser())
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
	}

	started, reason, err := vm_service.StartActivity(current.ID, vmModel.ActivityAttachingGPU)
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
		"id":     current.ID,
		"gpuId":  gpu.ID,
		"userId": auth.UserID,
	})

	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to create job: %s", err))
		return
	}

	context.JSONResponse(http.StatusOK, body.GpuAttached{
		ID:    current.ID,
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
// @Failure 400 {object} app.ErrorResponse
// @Failure 404 {object} app.ErrorResponse
// @Failure 423 {object} app.ErrorResponse
// @Failure 500 {object} app.ErrorResponse
// @Router /api/v1/vms/{vmId}/gpu [delete]
func DetachGPU(c *gin.Context) {
	context := app.NewContext(c)

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
