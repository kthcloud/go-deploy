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
	zoneModel "go-deploy/models/sys/zone"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	"go-deploy/service/job_service"
	"go-deploy/service/vm_service"
	"go-deploy/service/zone_service"
	"go-deploy/utils"
)

// List
// @Summary Get list of VMs
// @Description Get list of VMs
// @Tags VM
// @Accept  json
// @Produce  json
// @Param all query bool false "Get all"
// @Param userId query string false "Filter by user id"
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {array} body.VmRead
// @Failure 500 {object} sys.ErrorResponse
// @Router /vm [get]
func List(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.VmList
	if err := context.GinContext.Bind(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	vms, err := vm_service.ListAuth(requestQuery.All, requestQuery.UserID, auth, &requestQuery.Pagination)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if vms == nil {
		context.Ok([]interface{}{})
		return
	}

	dtoVMs := make([]body.VmRead, len(vms))
	for i, vm := range vms {
		connectionString, _ := vm_service.GetConnectionString(&vm)

		var gpuRead *body.GpuRead
		if vm.HasGPU() {
			gpu, err := vm_service.GetGpuByID(vm.GpuID, true)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to get gpu for vm when listing. details: %w", err))
			} else if gpu != nil {
				gpuDTO := gpu.ToDTO(true)
				gpuRead = &gpuDTO
			}
		}

		mapper, err := vm_service.GetExternalPortMapper(vm.ID)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to get external port mapper for vm when listing. details: %w", err))
			continue
		}

		dtoVMs[i] = vm.ToDTO(vm.StatusMessage, connectionString, gpuRead, mapper)
	}

	context.Ok(dtoVMs)
}

// Get
// @Summary Get VM by id
// @Description Get VM by id
// @Tags VM
// @Accept  json
// @Produce  json
// @Param vmId path string true "VM ID"
// @Success 200 {object} body.VmRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /vm/{vmId} [get]
func Get(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.VmGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	vm, err := vm_service.GetByIdAuth(requestURI.VmID, auth)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	connectionString, _ := vm_service.GetConnectionString(vm)
	var gpuRead *body.GpuRead
	if vm.HasGPU() {
		gpu, err := vm_service.GetGpuByID(vm.GpuID, true)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to get gpu for vm. details: %w", err))
		} else if gpu != nil {
			gpuDTO := gpu.ToDTO(true)
			gpuRead = &gpuDTO
		}
	}

	mapper, err := vm_service.GetExternalPortMapper(vm.ID)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to get external port mapper for vm. details: %w", err))
	}

	context.Ok(vm.ToDTO(vm.StatusMessage, connectionString, gpuRead, mapper))
}

// Create
// @Summary Create VM
// @Description Create VM
// @Tags VM
// @Accept  json
// @Produce  json
// @Param body body body.VmCreate true "VM body"
// @Success 200 {object} body.VmCreated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /vm [post]
func Create(c *gin.Context) {
	context := sys.NewContext(c)

	var requestBody body.VmCreate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	vm, err := vm_service.GetByName(requestBody.Name)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if vm != nil {
		context.UserError("VM already exists")
		return
	}

	if requestBody.Zone != nil {
		zone := zone_service.GetZone(*requestBody.Zone, zoneModel.ZoneTypeVM)
		if zone == nil {
			context.NotFound("Zone not found")
			return
		}
	}

	ok, reason, err := vm_service.CheckQuotaCreate(auth.UserID, &auth.GetEffectiveRole().Quotas, auth, requestBody)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if !ok {
		context.Forbidden(reason)
		return
	}

	vmID := uuid.New().String()
	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeCreateVM, map[string]interface{}{
		"id":      vmID,
		"ownerId": auth.UserID,
		"params":  requestBody,
	})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(body.VmCreated{
		ID:    vmID,
		JobID: jobID,
	})
}

// Delete
// @Summary Delete VM
// @Description Delete VM
// @Tags VM
// @Accept  json
// @Produce  json
// @Param vmId path string true "VM ID"
// @Success 200 {object} body.VmDeleted
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /vm/{vmId} [delete]
func Delete(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.VmDelete
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	current, err := vm_service.GetByIdAuth(requestURI.VmID, auth)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if current == nil {
		context.NotFound("VM not found")
		return
	}

	started, reason, err := vm_service.StartActivity(current.ID, vmModel.ActivityBeingDeleted)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if !started {
		context.Locked(reason)
		return
	}

	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeDeleteVM, map[string]interface{}{
		"id": current.ID,
	})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(body.VmDeleted{
		ID:    current.ID,
		JobID: jobID,
	})
}

// Update
// @Summary Update VM
// @Description Update VM
// @Tags VM
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param vmId path string true "VM ID"
// @Param body body body.VmUpdate true "VM update"
// @Success 200 {object} body.VmUpdated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /vm/{vmId} [put]
func Update(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.VmUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	var requestBody body.VmUpdate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	vm, err := vm_service.GetByIdAuth(requestURI.VmID, auth)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	if requestBody.GpuID != nil {
		updateGPU(&context, &requestBody, auth, vm)
		return
	}

	ok, reason, err := vm_service.CheckQuotaUpdate(auth.UserID, vm.ID, &auth.GetEffectiveRole().Quotas, auth, requestBody)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if !ok {
		context.Forbidden(reason)
		return
	}

	started, reason, err := vm_service.StartActivity(vm.ID, vmModel.ActivityBeingUpdated)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if !started {
		context.Locked(reason)
		return
	}

	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeUpdateVM, map[string]interface{}{
		"id":     vm.ID,
		"update": requestBody,
	})

	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(body.VmUpdated{
		ID:    vm.ID,
		JobID: jobID,
	})
}

func updateGPU(context *sys.ClientContext, requestBody *body.VmUpdate, auth *service.AuthInfo, vm *vmModel.VM) {
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

func detachGPU(context *sys.ClientContext, auth *service.AuthInfo, vm *vmModel.VM) {
	if vm.GpuID == "" {
		context.UserError("VM does not have a GPU attached")
		return
	}

	started, reason, err := vm_service.StartActivity(vm.ID, vmModel.ActivityDetachingGPU)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if !started {
		context.Locked(reason)
		return
	}

	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeDetachGpuFromVM, map[string]interface{}{
		"id":     vm.ID,
		"userId": auth.UserID,
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

func attachGPU(context *sys.ClientContext, requestBody *body.VmUpdate, auth *service.AuthInfo, vm *vmModel.VM) {
	if !auth.GetEffectiveRole().Permissions.UseGPUs {
		context.Forbidden("Tier does not include GPU access")
		return
	}

	var gpus []gpuModel.GPU
	if *requestBody.GpuID == "any" {
		if vm.HasGPU() {
			gpu, err := vm_service.GetGpuByID(vm.GpuID, false)
			if err != nil {
				context.ServerError(err, v1.InternalError)
				return
			}

			if gpu == nil {
				context.NotFound("GPU not found")
				return
			}

			if !gpu.Lease.IsExpired() {
				context.UserError("GPU lease not expired")
				return
			}

			gpus = []gpuModel.GPU{*gpu}
		} else {
			availableGpus, err := vm_service.GetAllAvailableGPU(auth.GetEffectiveRole().Permissions.UsePrivilegedGPUs)
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

		privilegedGPU, err := vm_service.IsGpuPrivileged(*requestBody.GpuID)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if privilegedGPU && !auth.GetEffectiveRole().Permissions.UsePrivilegedGPUs {
			context.NotFound("GPU not found")
			return
		}

		gpu, err := vm_service.GetGpuByID(*requestBody.GpuID, true)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if gpu == nil {
			context.NotFound("GPU not found")
			return
		}

		if vm.HasGPU() && vm.GpuID != *requestBody.GpuID {
			context.UserError("VM already has a GPU attached")
			return
		}

		if !gpu.Lease.IsExpired() {
			context.UserError("GPU lease not expired")
			return
		}

		hardwareAvailable, err := vm_service.IsGpuHardwareAvailable(gpu)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if !hardwareAvailable {
			context.ServerUnavailableError(fmt.Errorf("gpu hardware not available when attaching gpu to vm %s", vm.ID), v1.GpuNotAvailableErr)
			return
		}

		gpus = []gpuModel.GPU{*gpu}
	}

	if len(gpus) == 0 {
		context.ServerUnavailableError(fmt.Errorf("no available gpus when attaching gpu to vm %s", vm.ID), v1.NoAvailableGpuErr)
		return
	}

	// do this check to give a nice error to the user if the gpu cannot be attached
	// otherwise it will be silently ignored
	if len(gpus) == 1 {
		canStartOnHost, reason, err := vm_service.CanStartOnHost(vm.ID, gpus[0].Host)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if !canStartOnHost {
			if reason == "" {
				reason = "VM could not be started on host"
			}

			context.ServerUnavailableError(fmt.Errorf("vm %s could not be started on host %s when attaching gpu, details: %s", vm.ID, gpus[0].Host, reason), v1.HostNotAvailableErr)
			return
		}
	}

	gpuIds := make([]string, len(gpus))
	for i, gpu := range gpus {
		gpuIds[i] = gpu.ID
	}

	started, reason, err := vm_service.StartActivity(vm.ID, vmModel.ActivityAttachingGPU)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if !started {
		context.Locked(reason)
		return
	}

	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeAttachGpuToVM, map[string]interface{}{
		"id":            vm.ID,
		"gpuIds":        gpuIds,
		"userId":        auth.UserID,
		"leaseDuration": auth.GetEffectiveRole().Permissions.GpuLeaseDuration,
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
