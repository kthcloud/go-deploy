package v1_vm

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	jobModel "go-deploy/models/sys/job"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/job_service"
	"go-deploy/service/user_service"
	"go-deploy/service/vm_service"
	"log"
	"net/http"
)

func getAllVMs(context *app.ClientContext) {
	vms, err := vm_service.GetAll()
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	dtoVMs := make([]body.VmRead, len(vms))
	for i, vm := range vms {
		connectionString, _ := vm_service.GetConnectionString(&vm)

		var gpuRead *body.GpuRead
		if vm.GpuID != "" {
			gpu, err := vm_service.GetGpuByID(vm.GpuID, true)
			if err != nil {
				log.Printf("error getting gpu by id: %s", err)
			} else if gpu != nil {
				gpuDTO := gpu.ToDto(true)
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
// @Failure 500 {object} app.ErrorResponse
// @Router /api/v1/vm [get]
func GetList(c *gin.Context) {
	context := app.NewContext(c)

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

	if requestQuery.WantAll && auth.IsAdmin() {
		getAllVMs(&context)
		return
	}

	vms, _ := vm_service.GetByOwnerID(auth.UserID)
	if vms == nil {
		context.JSONResponse(200, []interface{}{})
		return
	}

	dtoVMs := make([]body.VmRead, len(vms))
	for i, vm := range vms {
		connectionString, _ := vm_service.GetConnectionString(&vm)

		var gpuRead *body.GpuRead
		if vm.GpuID != "" {
			gpu, err := vm_service.GetGpuByID(vm.GpuID, true)
			if err != nil {
				log.Printf("error getting gpu by id: %s", err)
			} else if gpu != nil {
				gpuDTO := gpu.ToDto(true)
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
// @Failure 400 {object} app.ErrorResponse
// @Failure 404 {object} app.ErrorResponse
// @Failure 500 {object} app.ErrorResponse
// @Router /api/v1/vm/{vmId} [get]
func Get(c *gin.Context) {
	context := app.NewContext(c)

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

	vm, err := vm_service.GetByID(auth.UserID, requestURI.VmID, auth.IsAdmin())
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
	if vm.GpuID != "" {
		gpu, err := vm_service.GetGpuByID(vm.GpuID, true)
		if err != nil {
			log.Printf("error getting gpu by id: %s", err)
		} else if gpu != nil {
			gpuDTO := gpu.ToDto(true)
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
// @Failure 400 {object} app.ErrorResponse
// @Failure 401 {object} app.ErrorResponse
// @Failure 404 {object} app.ErrorResponse
// @Failure 423 {object} app.ErrorResponse
// @Failure 500 {object} app.ErrorResponse
// @Router /api/v1/vm [post]
func Create(c *gin.Context) {
	context := app.NewContext(c)

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

	user, err := user_service.GetOrCreate(auth.UserID, auth.JwtToken.PreferredUsername, auth.Roles)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get user: %s", err))
		return
	}

	if user.ID != auth.UserID {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Created user id does not match auth user id"))
	}

	quota, err := user_service.GetQuotaByUserID(auth.UserID)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get quota: %s", err))
		return
	}

	if quota == nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Quota is not set for user"))
		return
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

	ok, reason, err := vm_service.CheckQuotaCreate(auth.UserID, quota, requestBody)
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
// @Failure 400 {object} app.ErrorResponse
// @Failure 401 {object} app.ErrorResponse
// @Failure 404 {object} app.ErrorResponse
// @Failure 423 {object} app.ErrorResponse
// @Failure 500 {object} app.ErrorResponse
// @Router /api/v1/vm/{vmId} [delete]
func Delete(c *gin.Context) {
	context := app.NewContext(c)

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

	current, err := vm_service.GetByID(auth.UserID, requestURI.VmID, auth.IsAdmin())
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
// @Failure 400 {object} app.ErrorResponse
// @Failure 401 {object} app.ErrorResponse
// @Failure 404 {object} app.ErrorResponse
// @Failure 423 {object} app.ErrorResponse
// @Failure 500 {object} app.ErrorResponse
// @Router /api/v1/vm/{vmId} [put]
func Update(c *gin.Context) {
	context := app.NewContext(c)

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

	current, err := vm_service.GetByID(auth.UserID, requestURI.VmID, auth.IsAdmin())
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, fmt.Sprintf("Failed to get vm: %s", err))
		return
	}

	if current == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("VM with id %s not found", requestURI.VmID))
		return
	}

	started, reason, err := vm_service.StartActivity(current.ID, vmModel.ActivityBeingUpdated)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to start activity: %s", err))
		return
	}

	if !started {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceNotReady, reason)
		return
	}

	if current.OwnerID != auth.UserID && !auth.IsAdmin() {
		context.ErrorResponse(http.StatusUnauthorized, status_codes.Error, "User is not allowed to update this resource")
		return
	}

	quota, err := user_service.GetQuotaByUserID(auth.UserID)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get quota: %s", err))
		return
	}

	if quota == nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Quota is not set for user"))
		return
	}

	ok, reason, err := vm_service.CheckQuotaUpdate(auth.UserID, quota, requestBody)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to check quota: %s", err))
		return
	}

	if !ok {
		context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceValidationFailed, reason)
		return
	}

	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeUpdateVM, map[string]interface{}{
		"id":     current.ID,
		"update": requestBody,
	})

	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to create job: %s", err))
		return
	}

	context.JSONResponse(http.StatusOK, body.VmUpdated{
		ID:    current.ID,
		JobID: jobID,
	})
}
