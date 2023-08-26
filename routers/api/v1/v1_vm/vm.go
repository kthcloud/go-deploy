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
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	"go-deploy/service/job_service"
	"go-deploy/service/vm_service"
	"go-deploy/service/zone_service"
	"log"
	"net/http"
)

func getAllVMs(context *sys.ClientContext, auth *service.AuthInfo) {
	vms, err := vm_service.GetAllAuth(auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	dtoVMs := make([]body.VmRead, len(vms))
	for i, vm := range vms {
		connectionString, _ := vm_service.GetConnectionString(&vm)

		var gpuRead *body.GpuRead
		if vm.HasGPU() {
			gpu, err := vm_service.GetGpuByID(vm.GpuID, true)
			if err != nil {
				log.Printf("error getting gpu by id: %s", err)
			} else if gpu != nil {
				gpuDTO := gpu.ToDTO(true)
				gpuRead = &gpuDTO
			}
		}

		mapper, err := vm_service.GetExternalPortMapper(vm.ID)
		if err != nil {
			log.Printf("error getting external port mapper: %s", err)
			continue
		}

		dtoVMs[i] = vm.ToDTO(vm.StatusMessage, connectionString, gpuRead, mapper)
	}

	context.JSONResponse(http.StatusOK, dtoVMs)
}

// GetList
// @Summary Get list of VMs
// @Description Get list of VMs
// @Tags VM
// @Accept  json
// @Produce  json
// @Param wantAll query bool false "Want all VMs"
// @Success 200 {array} body.VmRead
// @Failure 500 {object} sys.ErrorResponse
// @Router /api/v1/vm [get]
func GetList(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.VmList
	if err := context.GinContext.Bind(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	if requestQuery.WantAll && auth.IsAdmin {
		getAllVMs(&context, auth)
		return
	}

	vms, _ := vm_service.GetByOwnerIdAuth(auth.UserID, auth)
	if vms == nil {
		context.JSONResponse(200, []interface{}{})
		return
	}

	dtoVMs := make([]body.VmRead, len(vms))
	for i, vm := range vms {
		connectionString, _ := vm_service.GetConnectionString(&vm)

		var gpuRead *body.GpuRead
		if vm.HasGPU() {
			gpu, err := vm_service.GetGpuByID(vm.GpuID, true)
			if err != nil {
				log.Printf("error getting gpu by id: %s", err)
			} else if gpu != nil {
				gpuDTO := gpu.ToDTO(true)
				gpuRead = &gpuDTO
			}
		}

		mapper, err := vm_service.GetExternalPortMapper(vm.ID)
		if err != nil {
			log.Printf("error getting external port mapper: %s", err)
			continue
		}

		dtoVMs[i] = vm.ToDTO(vm.StatusMessage, connectionString, gpuRead, mapper)
	}

	context.JSONResponse(200, dtoVMs)
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
// @Router /api/v1/vm/{vmId} [get]
func Get(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.VmGet
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err.Error()))
		return
	}

	vm, err := vm_service.GetByIdAuth(requestURI.VmID, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get vm: %s", err.Error()))
		return
	}

	if vm == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("VM with id %s not found", requestURI.VmID))
		return
	}

	connectionString, _ := vm_service.GetConnectionString(vm)
	var gpuRead *body.GpuRead
	if vm.HasGPU() {
		gpu, err := vm_service.GetGpuByID(vm.GpuID, true)
		if err != nil {
			log.Printf("error getting gpu by id: %s", err)
		} else if gpu != nil {
			gpuDTO := gpu.ToDTO(true)
			gpuRead = &gpuDTO
		}
	}

	mapper, err := vm_service.GetExternalPortMapper(vm.ID)
	if err != nil {
		log.Printf("error getting external port mapper: %s", err)
	}

	context.JSONResponse(200, vm.ToDTO(vm.StatusMessage, connectionString, gpuRead, mapper))
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
// @Router /api/v1/vm [post]
func Create(c *gin.Context) {
	context := sys.NewContext(c)

	var requestBody body.VmCreate
	if err := context.GinContext.BindJSON(&requestBody); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	if requestBody.Zone != nil {
		zone := zone_service.GetZone(*requestBody.Zone, zoneModel.ZoneTypeVM)
		if zone == nil {
			context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, "Zone not found")
			return
		}
	}

	exists, vm, err := vm_service.Exists(requestBody.Name)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if exists {
		if vm.OwnerID != auth.UserID {
			context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceNotCreated, "Resource already exists")
			return
		}

		started, reason, err := vm_service.StartActivity(vm.ID, vmModel.ActivityBeingCreated)
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to start activity: %s", err))
			return
		}

		if !started {
			context.ErrorResponse(http.StatusLocked, status_codes.ResourceNotReady, reason)
			return
		}

		if vm.BeingDeleted() {
			context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingDeleted, "Resource is currently being deleted")
			return
		}

		jobID := uuid.New().String()
		err = job_service.Create(jobID, auth.UserID, jobModel.TypeCreateVM, map[string]interface{}{
			"id":      vm.ID,
			"ownerId": auth.UserID,
			"params":  requestBody,
		})
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to create job: %s", err))
			return
		}

		context.JSONResponse(http.StatusCreated, body.VmCreated{
			ID:    vm.ID,
			JobID: jobID,
		})
		return
	}

	ok, reason, err := vm_service.CheckQuotaCreate(auth.UserID, &auth.GetEffectiveRole().Quotas, requestBody)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to check quota: %s", err))
		return
	}

	if !ok {
		context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceValidationFailed, reason)
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
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to create job: %s", err))
		return
	}

	context.JSONResponse(http.StatusCreated, body.VmCreated{
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
// @Router /api/v1/vm/{vmId} [delete]
func Delete(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.VmDelete
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	current, err := vm_service.GetByIdAuth(requestURI.VmID, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, fmt.Sprintf("VM with id %s not found", requestURI.VmID))
		return
	}

	if current == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, "Resource not found")
		return
	}

	started, reason, err := vm_service.StartActivity(current.ID, vmModel.ActivityBeingDeleted)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to start activity: %s", err))
		return
	}

	if !started {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceNotReady, reason)
		return
	}

	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeDeleteVM, map[string]interface{}{
		"name": current.Name,
	})
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to create job: %s", err))
		return
	}

	context.JSONResponse(http.StatusOK, body.VmDeleted{
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
// @Param Authorization header string true "With the bearer started"
// @Param vmId path string true "VM ID"
// @Param body body body.VmUpdate true "VM update"
// @Success 200 {object} body.VmUpdated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /api/v1/vm/{vmId} [put]
func Update(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.VmUpdate
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	var requestBody body.VmUpdate
	if err := context.GinContext.BindJSON(&requestBody); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	vm, err := vm_service.GetByIdAuth(requestURI.VmID, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, fmt.Sprintf("Failed to get vm: %s", err))
		return
	}

	if vm == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("VM with id %s not found", requestURI.VmID))
		return
	}

	if requestBody.GpuID != nil {
		updateGPU(&context, &requestBody, auth, vm)
		return
	}

	ok, reason, err := vm_service.CheckQuotaUpdate(auth.UserID, vm.ID, &auth.GetEffectiveRole().Quotas, requestBody)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to check quota: %s", err))
		return
	}

	if !ok {
		context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceValidationFailed, reason)
		return
	}

	started, reason, err := vm_service.StartActivity(vm.ID, vmModel.ActivityBeingUpdated)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to start activity: %s", err))
		return
	}

	if !started {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceNotReady, reason)
		return
	}

	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeUpdateVM, map[string]interface{}{
		"id":     vm.ID,
		"update": requestBody,
	})

	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to create job: %s", err))
		return
	}

	context.JSONResponse(http.StatusOK, body.VmUpdated{
		ID:    vm.ID,
		JobID: jobID,
	})
}

func updateGPU(context *sys.ClientContext, requestBody *body.VmUpdate, auth *service.AuthInfo, vm *vmModel.VM) {
	decodedGpuID, decodeErr := decodeGpuID(*requestBody.GpuID)
	if decodeErr != nil {
		context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceValidationFailed, "Invalid GPU ID")
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
		context.ErrorResponse(http.StatusNotModified, status_codes.ResourceNotUpdated, "VM does not have a GPU attached")
		return
	}

	started, reason, err := vm_service.StartActivity(vm.ID, vmModel.ActivityDetachingGPU)
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
		"id":     vm.ID,
		"userId": auth.UserID,
	})
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to create job: %s", err))
		return
	}

	context.JSONResponse(http.StatusOK, body.GpuDetached{
		ID:    vm.ID,
		JobID: jobID,
	})
}

func attachGPU(context *sys.ClientContext, requestBody *body.VmUpdate, auth *service.AuthInfo, vm *vmModel.VM) {
	if !auth.GetEffectiveRole().Permissions.UseGPUs {
		context.ErrorResponse(http.StatusForbidden, status_codes.Error, "Tier does not include GPU access")
		return
	}

	var gpus []gpuModel.GPU
	if *requestBody.GpuID == "any" {
		if vm.HasGPU() {
			gpu, err := vm_service.GetGpuByID(vm.GpuID, false)
			if err != nil {
				context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get gpu: %s", err))
				return
			}

			if gpu == nil {
				context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("GPU with id %s not found", vm.GpuID))
				return
			}

			if !gpu.Lease.IsExpired() {
				context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceNotCreated, "GPU lease not expired")
				return
			}

			gpus = []gpuModel.GPU{*gpu}
		} else {
			availableGpus, err := vm_service.GetAllAvailableGPU(auth.GetEffectiveRole().Permissions.UsePrivilegedGPUs)
			if err != nil {
				context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get available GPUs: %s", err))
				return
			}

			if availableGpus == nil {
				context.ErrorResponse(http.StatusNotModified, status_codes.ResourceNotAvailable, "No available GPUs")
				return
			}

			gpus = availableGpus
		}
	} else {
		if !auth.GetEffectiveRole().Permissions.ChooseGPU {
			context.ErrorResponse(http.StatusForbidden, status_codes.Error, "Tier does not include GPU selection")
			return
		}

		privilegedGPU, err := vm_service.IsGpuPrivileged(*requestBody.GpuID)
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get gpu: %s", err))
			return
		}

		if privilegedGPU && !auth.GetEffectiveRole().Permissions.UsePrivilegedGPUs {
			context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("GPU with id %s not found", *requestBody.GpuID))
			return
		}

		gpu, err := vm_service.GetGpuByID(*requestBody.GpuID, false)
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get gpu: %s", err))
			return
		}

		if gpu == nil {
			context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("GPU with id %s not found", *requestBody.GpuID))
			return
		}

		if vm.HasGPU() && vm.GpuID != *requestBody.GpuID {
			context.ErrorResponse(http.StatusBadRequest, status_codes.Error, "VM already has a GPU attached")
			return
		}

		if !gpu.Lease.IsExpired() {
			context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceNotCreated, "GPU lease not expired")
			return
		}

		hardwareAvailable, err := vm_service.IsGpuHardwareAvailable(gpu)
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to check if GPU is available: %s", err))
			return
		}

		if !hardwareAvailable {
			context.ErrorResponse(http.StatusNotModified, status_codes.ResourceNotAvailable, "GPU not available")
			return
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
				reason = "VM could not be started on host"
			}

			context.ErrorResponse(http.StatusNotModified, status_codes.ResourceNotUpdated, reason)
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
		"id":            vm.ID,
		"gpuIds":        gpuIds,
		"userId":        auth.UserID,
		"leaseDuration": auth.GetEffectiveRole().Permissions.GpuLeaseDuration,
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
