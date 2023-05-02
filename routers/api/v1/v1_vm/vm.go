package v1_vm

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	jobModel "go-deploy/models/job"
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
				gpuDTO := gpu.ToDto()
				gpuRead = &gpuDTO
			}
		}

		dtoVMs[i] = vm.ToDTO(vm.StatusMessage, connectionString, gpuRead)
	}

	context.JSONResponse(http.StatusOK, dtoVMs)
}

func GetList(c *gin.Context) {
	context := app.NewContext(c)

	var requestQuery query.VmList
	if err := context.GinContext.BindQuery(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(&requestQuery, err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err.Error()))
		return
	}

	if requestQuery.WantAll && auth.IsAdmin {
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
				gpuDTO := gpu.ToDto()
				gpuRead = &gpuDTO
			}
		}

		dtoVMs[i] = vm.ToDTO(vm.StatusMessage, connectionString, gpuRead)
	}

	context.JSONResponse(200, dtoVMs)
}

func Get(c *gin.Context) {
	context := app.NewContext(c)

	var requestURI uri.VmGet
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(&requestURI, err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err.Error()))
		return
	}

	vm, err := vm_service.GetByID(auth.UserID, requestURI.VmID, auth.IsAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
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
			gpuDTO := gpu.ToDto()
			gpuRead = &gpuDTO
		}
	}

	context.JSONResponse(200, vm.ToDTO(vm.StatusMessage, connectionString, gpuRead))
}

func Create(c *gin.Context) {
	context := app.NewContext(c)

	var requestBody body.VmCreate
	if err := context.GinContext.BindJSON(&requestBody); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(&requestBody, err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	userInfo, err := user_service.GetOrCreate(auth.JwtToken)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if userInfo.ID != auth.UserID {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Created user id does not match auth user id"))
	}

	if userInfo.VmQuota == 0 {
		context.ErrorResponse(http.StatusUnauthorized, status_codes.Error, "User is not allowed to create vms")
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
		if vm.BeingDeleted {
			context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingDeleted, "Resource is currently being deleted")
			return
		}

		jobID := uuid.New().String()
		err = job_service.Create(jobID, auth.UserID, jobModel.TypeCreateVM, map[string]interface{}{
			"id":           vm.ID,
			"name":         requestBody.Name,
			"sshPublicKey": requestBody.SshPublicKey,
			"ownerId":      auth.UserID,
		})
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
			return
		}

		context.JSONResponse(http.StatusCreated, body.VmCreated{
			ID:    vm.ID,
			JobID: jobID,
		})
		return
	}

	vmCount, err := vm_service.GetCount(auth.UserID)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if vmCount >= userInfo.VmQuota {
		context.ErrorResponse(http.StatusUnauthorized, status_codes.Error, fmt.Sprintf("User is not allowed to create more than %d vms", userInfo.VmQuota))
		return
	}

	vmID := uuid.New().String()
	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeCreateVM, map[string]interface{}{
		"id":           vmID,
		"name":         requestBody.Name,
		"sshPublicKey": requestBody.SshPublicKey,
		"ownerId":      auth.UserID,
	})

	context.JSONResponse(http.StatusCreated, body.VmCreated{
		ID:    vmID,
		JobID: jobID,
	})
}

func Delete(c *gin.Context) {
	context := app.NewContext(c)

	var requestURI uri.VmDelete
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
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, fmt.Sprintf("VM with id %s not found", requestURI.VmID))
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

	if !current.BeingDeleted {
		_ = vm_service.MarkBeingDeleted(current.ID)
	}

	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeDeleteVM, map[string]interface{}{
		"name": current.Name,
	})

	context.JSONResponse(http.StatusOK, body.VmDeleted{
		ID:    current.ID,
		JobID: jobID,
	})
}

func Update(c *gin.Context) {
	context := app.NewContext(c)

	var requestURI uri.VmUpdate
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(&requestURI, err))
		return
	}

	var requestBody body.VmUpdate
	if err := context.GinContext.BindJSON(&requestBody); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(&requestBody, err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	current, err := vm_service.GetByID(auth.UserID, requestURI.VmID, auth.IsAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, fmt.Sprintf("Failed to get vm: %s", err))
		return
	}

	if current == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("VM with id %s not found", requestURI.VmID))
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

	if current.OwnerID != auth.UserID && !auth.IsAdmin {
		context.ErrorResponse(http.StatusUnauthorized, status_codes.Error, "User is not allowed to update this vm")
		return
	}

	err = vm_service.Update(current.ID, &requestBody)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update vm: %s", err))
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
