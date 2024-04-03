package v1

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/dto/v1/body"
	"go-deploy/dto/v1/query"
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/sys"
	"go-deploy/service"
	"go-deploy/service/clients"
	sErrors "go-deploy/service/errors"
	v12 "go-deploy/service/v1/utils"
	"go-deploy/service/v1/vms/opts"
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
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	gpus, err := service.V1(auth).VMs().ListGPUs(opts.ListGpuOpts{
		Pagination:    v12.GetOrDefaultPagination(requestQuery.Pagination),
		Zone:          requestQuery.Zone,
		AvailableGPUs: requestQuery.OnlyShowAvailable,
	})
	if err != nil {
		context.ServerError(err, InternalError)
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
//
// Auth is assumed to be set in the V1 client.
func updateGPU(context *sys.ClientContext, requestBody *body.VmUpdate, deployV1 clients.V1, vm *model.VM) {
	decodedGpuID, decodeErr := decodeGpuID(*requestBody.GpuID)
	if decodeErr != nil {
		context.UserError("Invalid GPU ID")
		return
	}

	requestBody.GpuID = &decodedGpuID

	if *requestBody.GpuID == "" {
		detachGPU(context, deployV1, vm)
		return
	} else {
		attachGPU(context, requestBody, deployV1, vm)
		return
	}
}

// detachGPU is an alternate entrypoint for UpdateVM that allows a user to detach a GPU from a VM
// It is called if an empty GPU ID is passed in the request body for Update.
//
// Auth is assumed to be set in the V1 client.
func detachGPU(context *sys.ClientContext, deployV1 clients.V1, vm *model.VM) {
	err := deployV1.VMs().StartActivity(vm.ID, model.ActivityDetachingGPU)
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

		context.ServerError(err, InternalError)
		return
	}

	jobID := uuid.New().String()
	err = deployV1.Jobs().Create(jobID, deployV1.Auth().UserID, model.JobDetachGPU, version.V1, map[string]interface{}{
		"id": vm.ID,
	})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	context.Ok(body.GpuDetached{
		ID:    vm.ID,
		JobID: jobID,
	})
}

// attachGPU is an alternate entrypoint for UpdateVM that allows a user to attach a GPU to a VM
// It is called if a GPU ID is passed in the request body for Update.
func attachGPU(context *sys.ClientContext, requestBody *body.VmUpdate, deployV1 clients.V1, vm *model.VM) {
	if !deployV1.Auth().GetEffectiveRole().Permissions.UseGPUs {
		context.Forbidden("Tier does not include GPU access")
		return
	}

	currentGPU, err := deployV1.VMs().GetGpuByVM(vm.ID)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	var gpus []model.GPU
	if *requestBody.GpuID == "any" {
		if currentGPU != nil {
			if !currentGPU.Lease.IsExpired() {
				context.UserError("GPU lease not expired")
				return
			}

			gpus = []model.GPU{*currentGPU}
		} else {
			availableGpus, err := deployV1.VMs().ListGPUs(opts.ListGpuOpts{
				AvailableGPUs: true,
				Zone:          &vm.Zone,
			})
			if err != nil {
				context.ServerError(err, InternalError)
				return
			}

			if availableGpus == nil {
				context.ServerUnavailableError(fmt.Errorf("no available gpus when attaching gpu_repo to vm %s", vm.ID), NoAvailableGpuErr)
				return
			}

			gpus = availableGpus
		}
	} else {
		if !deployV1.Auth().GetEffectiveRole().Permissions.ChooseGPU {
			context.Forbidden("Tier does not include GPU selection")
			return
		}

		privilegedGPU, err := deployV1.VMs().IsGpuPrivileged(*requestBody.GpuID)
		if err != nil {
			context.ServerError(err, InternalError)
			return
		}

		if privilegedGPU && !deployV1.Auth().GetEffectiveRole().Permissions.UsePrivilegedGPUs {
			context.NotFound("GPU not found")
			return
		}

		requestedGPU, err := deployV1.VMs().GetGPU(*requestBody.GpuID, opts.GetGpuOpts{
			Zone: &vm.Zone,
		})
		if err != nil {
			context.ServerError(err, InternalError)
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

		err = deployV1.VMs().CheckGpuHardwareAvailable(requestedGPU.ID)
		if err != nil {
			switch {
			case errors.Is(err, sErrors.HostNotAvailableErr):
				context.ServerUnavailableError(fmt.Errorf("host not available when attaching gpu_repo to vm %s. details: %w", vm.ID, err), HostNotAvailableErr)
			default:
				context.ServerError(err, InternalError)
			}
			return
		}

		gpus = []model.GPU{*requestedGPU}
	}

	if len(gpus) == 0 {
		context.ServerUnavailableError(fmt.Errorf("no available gpus when attaching gpu_repo to vm %s", vm.ID), NoAvailableGpuErr)
		return
	}

	// Do this check to give a nice error to the user if the gpu_repo cannot be attached
	// Otherwise it will be silently ignored
	if len(gpus) == 1 {
		if err := deployV1.VMs().CheckSuitableHost(vm.ID, gpus[0].Host, gpus[0].Zone); err != nil {
			switch {
			case errors.Is(err, sErrors.HostNotAvailableErr):
				context.ServerUnavailableError(fmt.Errorf("host not available when attaching gpu_repo to vm %s. details: %w", vm.ID, err), HostNotAvailableErr)
			case errors.Is(err, sErrors.VmTooLargeErr):
				tooLargeErr := VmTooLargeForHostErr
				caps, err := deployV1.VMs().GetCloudStackHostCapabilities(gpus[0].Host, vm.Zone)
				if err == nil && caps != nil {
					tooLargeErr = MakeVmToLargeForHostErr(caps.CpuCoresTotal-caps.CpuCoresUsed, caps.RamTotal-caps.RamAllocated)
				}
				context.ServerUnavailableError(fmt.Errorf("vm %s too large when attaching gpu_repo", vm.ID), tooLargeErr)
			case errors.Is(err, sErrors.VmNotCreatedErr):
				context.ServerUnavailableError(fmt.Errorf("vm %s not created when attaching gpu_repo to vm %s. details: %w", vm.ID, vm.ID, err), VmNotReadyErr)
			default:
				context.ServerError(err, InternalError)
			}
			return
		}
	}

	gpuIds := make([]string, len(gpus))
	for i, gpu := range gpus {
		gpuIds[i] = gpu.ID
	}

	err = deployV1.VMs().StartActivity(vm.ID, model.ActivityAttachingGPU)
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

		context.ServerError(err, InternalError)
		return
	}

	var leaseDuration float64
	if requestBody.NoLeaseEnd != nil && *requestBody.NoLeaseEnd && deployV1.Auth().IsAdmin {
		leaseDuration = -1
	} else {
		leaseDuration = deployV1.Auth().GetEffectiveRole().Quotas.GpuLeaseDuration
	}

	jobID := uuid.New().String()
	err = deployV1.Jobs().Create(jobID, deployV1.Auth().UserID, model.JobAttachGPU, version.V1, map[string]interface{}{
		"id":            vm.ID,
		"gpuIds":        gpuIds,
		"userId":        deployV1.Auth().UserID,
		"leaseDuration": leaseDuration,
	})

	if err != nil {
		context.ServerError(err, InternalError)
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
