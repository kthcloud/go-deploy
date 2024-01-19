package v1_vm

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	gpuModels "go-deploy/models/sys/gpu"
	jobModels "go-deploy/models/sys/job"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/job_service"
	"go-deploy/service/vm_service"
	"go-deploy/service/vm_service/client"
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

	gpus, err := vm_service.New().WithAuth(auth).ListGPUs(&client.ListGpuOptions{
		Pagination:    service.GetOrDefault(requestQuery.Pagination),
		Zone:          requestQuery.Zone,
		AvailableGPUs: requestQuery.OnlyShowAvailable,
	})
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

// updateGPU is an alternate entrypoint for UpdateVM that allows a user to attach or detach a GPU to a VM
// It is called if GPU ID is not nil in the request body for Update.
func updateGPU(context *sys.ClientContext, requestBody *body.VmUpdate, auth *service.AuthInfo, vm *vmModels.VM) {
	decodedGpuID, decodeErr := decodeGpuID(*requestBody.GpuID)
	if decodeErr != nil {
		context.UserError("Invalid GPU ID")
		return
	}

	requestBody.GpuID = &decodedGpuID

	if *requestBody.GpuID == "" {
		detachGPU(context, auth, vm)
		return
	} else {
		attachGPU(context, requestBody, auth, vm)
		return
	}
}

// detachGPU is an alternate entrypoint for UpdateVM that allows a user to detach a GPU from a VM
// It is called if an empty GPU ID is passed in the request body for Update.
func detachGPU(context *sys.ClientContext, auth *service.AuthInfo, vm *vmModels.VM) {
	if !vm.HasGPU() {
		context.UserError("VM does not have a GPU attached")
		return
	}

	vsc := vm_service.New().WithAuth(auth)

	err := vsc.StartActivity(vm.ID, vmModels.ActivityDetachingGPU)
	if err != nil {
		var failedToStartActivityErr sErrors.FailedToStartActivityError
		if errors.As(err, &failedToStartActivityErr) {
			context.Locked(failedToStartActivityErr.Error())
			return
		}

		if errors.Is(err, sErrors.VmNotFoundErr) {
			context.NotFound("Deployment not found")
			return
		}

		context.ServerError(err, v1.InternalError)
		return
	}

	jobID := uuid.New().String()
	err = job_service.New().Create(jobID, auth.UserID, jobModels.TypeDetachGPU, map[string]interface{}{
		"id": vm.ID,
	})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(body.GpuDetached{
		ID:    vm.ID,
		JobID: jobID,
	})
}

// attachGPU is an alternate entrypoint for UpdateVM that allows a user to attach a GPU to a VM
// It is called if a GPU ID is passed in the request body for Update.
func attachGPU(context *sys.ClientContext, requestBody *body.VmUpdate, auth *service.AuthInfo, vm *vmModels.VM) {
	if !auth.GetEffectiveRole().Permissions.UseGPUs {
		context.Forbidden("Tier does not include GPU access")
		return
	}

	vsc := vm_service.New().WithAuth(auth)
	currentGPU := vm.GetGpu()

	var gpus []gpuModels.GPU
	if *requestBody.GpuID == "any" {
		if currentGPU != nil {
			if !currentGPU.Lease.IsExpired() {
				context.UserError("GPU lease not expired")
				return
			}

			gpus = []gpuModels.GPU{*currentGPU}
		} else {
			availableGpus, err := vsc.ListGPUs(&client.ListGpuOptions{
				AvailableGPUs: true,
				Zone:          &vm.Zone,
			})
			if err != nil {
				context.ServerError(err, v1.InternalError)
				return
			}

			if availableGpus == nil {
				context.ServerUnavailableError(fmt.Errorf("no available gpus when attaching gpu to vm %s", vm.ID), v1.NoAvailableGpuErr)
				return
			}

			gpus = availableGpus
		}
	} else {
		if !auth.GetEffectiveRole().Permissions.ChooseGPU {
			context.Forbidden("Tier does not include GPU selection")
			return
		}

		privilegedGPU, err := vsc.IsGpuPrivileged(*requestBody.GpuID)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if privilegedGPU && !auth.GetEffectiveRole().Permissions.UsePrivilegedGPUs {
			context.NotFound("GPU not found")
			return
		}

		requestedGPU, err := vsc.GetGPU(*requestBody.GpuID, &client.GetGpuOptions{
			Zone: &vm.Zone,
		})
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if requestedGPU == nil {
			context.NotFound("GPU not found")
			return
		}

		if currentGPU != nil && currentGPU.ID != requestedGPU.ID {
			context.UserError("VM already has a GPU attached")
			return
		}

		if !requestedGPU.Lease.IsExpired() {
			context.UserError("GPU lease not expired")
			return
		}

		err = vsc.CheckGpuHardwareAvailable(requestedGPU.ID)
		if err != nil {
			switch {
			case errors.Is(err, sErrors.HostNotAvailableErr):
				context.ServerUnavailableError(fmt.Errorf("host not available when attaching gpu to vm %s. details: %w", vm.ID, err), v1.HostNotAvailableErr)
			default:
				context.ServerError(err, v1.InternalError)
			}
			return
		}

		gpus = []gpuModels.GPU{*requestedGPU}
	}

	if len(gpus) == 0 {
		context.ServerUnavailableError(fmt.Errorf("no available gpus when attaching gpu to vm %s", vm.ID), v1.NoAvailableGpuErr)
		return
	}

	// do this check to give a nice error to the user if the gpu cannot be attached
	// otherwise it will be silently ignored
	if len(gpus) == 1 {
		if err := vsc.CheckSuitableHost(vm.ID, gpus[0].Host, gpus[0].Zone); err != nil {
			switch {
			case errors.Is(err, sErrors.HostNotAvailableErr):
				context.ServerUnavailableError(fmt.Errorf("host not available when attaching gpu to vm %s. details: %w", vm.ID, err), v1.HostNotAvailableErr)
			case errors.Is(err, sErrors.VmTooLargeErr):
				tooLargeErr := v1.VmTooLargeForHostErr
				caps, err := vm_service.GetCloudStackHostCapabilities(gpus[0].Host, vm.Zone)
				if err == nil && caps != nil {
					tooLargeErr = v1.MakeVmToLargeForHostErr(caps.CpuCoresTotal-caps.CpuCoresUsed, caps.RamTotal-caps.RamUsed)
				}
				context.ServerUnavailableError(fmt.Errorf("vm %s too large when attaching gpu", vm.ID), tooLargeErr)
			case errors.Is(err, sErrors.VmNotCreatedErr):
				context.ServerUnavailableError(fmt.Errorf("vm %s not created when attaching gpu to vm %s. details: %w", vm.ID, vm.ID, err), v1.VmNotReadyErr)
			default:
				context.ServerError(err, v1.InternalError)
			}
			return
		}
	}

	gpuIds := make([]string, len(gpus))
	for i, gpu := range gpus {
		gpuIds[i] = gpu.ID
	}

	err := vsc.StartActivity(vm.ID, vmModels.ActivityAttachingGPU)
	if err != nil {
		var failedToStartActivityErr sErrors.FailedToStartActivityError
		if errors.As(err, &failedToStartActivityErr) {
			context.Locked(failedToStartActivityErr.Error())
			return
		}

		if errors.Is(err, sErrors.VmNotFoundErr) {
			context.NotFound("Deployment not found")
			return
		}

		context.ServerError(err, v1.InternalError)
		return
	}

	jobID := uuid.New().String()
	err = job_service.New().Create(jobID, auth.UserID, jobModels.TypeAttachGPU, map[string]interface{}{
		"id":            vm.ID,
		"gpuIds":        gpuIds,
		"userId":        auth.UserID,
		"leaseDuration": auth.GetEffectiveRole().Quotas.GpuLeaseDuration,
	})

	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(body.GpuAttached{
		ID:    vm.ID,
		JobID: jobID,
	})
}

// decodeGpuID is a helper function that decodes a base64 encoded GPU ID
func decodeGpuID(gpuID string) (string, error) {
	if gpuID == "any" {
		return gpuID, nil
	}

	res, err := base64.StdEncoding.DecodeString(gpuID)
	if err != nil {
		return "", err
	}
	return string(res), nil
}
