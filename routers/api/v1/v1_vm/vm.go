package v1_vm

import (
	"fmt"
	"go-deploy/models/dto"
	jobModel "go-deploy/models/job"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/validator"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/job_service"
	"go-deploy/service/user_info_service"
	"go-deploy/service/vm_service"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

func getAllVMs(context *app.ClientContext) {
	vms, err := vm_service.GetAll()
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	dtoVMs := make([]dto.VmRead, len(vms))
	for i, vm := range vms {
		connectionString, _ := vm_service.GetConnectionString(&vm)

		var gpuRead *dto.GpuRead
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

	rules := validator.MapData{
		"all": []string{"bool"},
	}

	validationErrors := context.ValidateQueryParams(&rules)
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

	isAdmin := v1.IsAdmin(&context)
	wantAll, _ := strconv.ParseBool(context.GinContext.Query("all"))
	if wantAll && isAdmin {
		getAllVMs(&context)
		return
	}

	vms, _ := vm_service.GetByOwnerID(userID)
	if vms == nil {
		context.JSONResponse(200, []interface{}{})
		return
	}

	dtoVMs := make([]dto.VmRead, len(vms))
	for i, vm := range vms {
		connectionString, _ := vm_service.GetConnectionString(&vm)

		var gpuRead *dto.GpuRead
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

	rules := validator.MapData{
		"vmId": []string{"required", "uuid_v4"},
	}

	validationErrors := context.ValidateParams(&rules)

	if len(validationErrors) > 0 {
		context.ResponseValidationError(validationErrors)
		return
	}

	token, err := context.GetKeycloakToken()
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
	}
	vmID := context.GinContext.Param("vmId")
	userID := token.Sub
	isAdmin := v1.IsAdmin(&context)

	vm, err := vm_service.GetByID(userID, vmID, isAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if vm == nil {
		context.NotFound()
		return
	}

	connectionString, _ := vm_service.GetConnectionString(vm)
	var gpuRead *dto.GpuRead
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

	bodyRules := validator.MapData{
		"name": []string{
			"required",
			"regex:^[a-zA-Z]([a-zA-Z0-9-]*[a-zA-Z0-9])?([a-zA-Z]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$",
			"min:3",
			"max:30",
		},
		"sshPublicKey": []string{
			"required",
		},
	}

	messages := validator.MapData{
		"name": []string{
			"required:Name is required",
			"regexp:Name must follow RFC 1035 and must not include any dots",
			"min:Name must be between 3-30 characters",
			"max:Name must be between 3-30 characters",
		},
		"sshPublicKey": []string{
			"required:SSH public key is required",
		},
	}

	var requestBody dto.VmCreate
	validationErrors := context.ValidateJSONCustomMessages(&bodyRules, &messages, &requestBody)
	if len(validationErrors) > 0 {
		context.ResponseValidationError(validationErrors)
		return
	}

	token, err := context.GetKeycloakToken()
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	userInfo, err := user_info_service.GetByToken(token)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if userInfo.VmQuota == 0 {
		context.ErrorResponse(http.StatusUnauthorized, status_codes.Error, "User is not allowed to create vms")
		return
	}

	userID := token.Sub
	_ = v1.IsAdmin(&context)

	validKey := isValidSshPublicKey(requestBody.SshPublicKey)
	if !validKey {
		context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceValidationFailed, "Invalid SSH public key")
		return
	}

	exists, vm, err := vm_service.Exists(requestBody.Name)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if exists {
		if vm.OwnerID != userID {
			context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceNotCreated, "Resource already exists")
			return
		}
		if vm.BeingDeleted {
			context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingDeleted, "Resource is currently being deleted")
			return
		}

		jobID := uuid.New().String()
		err = job_service.Create(jobID, userID, jobModel.TypeCreateVM, map[string]interface{}{
			"id":           vm.ID,
			"name":         requestBody.Name,
			"sshPublicKey": requestBody.SshPublicKey,
			"ownerId":      userID,
		})
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
			return
		}

		context.JSONResponse(http.StatusCreated, dto.VmCreated{
			ID:    vm.ID,
			JobID: jobID,
		})
		return
	}

	vmCount, err := vm_service.GetCount(userID)
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
	err = job_service.Create(jobID, userID, jobModel.TypeCreateVM, map[string]interface{}{
		"id":           vmID,
		"name":         requestBody.Name,
		"sshPublicKey": requestBody.SshPublicKey,
		"ownerId":      userID,
	})

	context.JSONResponse(http.StatusCreated, dto.VmCreated{
		ID:    vmID,
		JobID: jobID,
	})
}

func Delete(c *gin.Context) {
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

	if !current.BeingDeleted {
		_ = vm_service.MarkBeingDeleted(current.ID)
	}

	jobID := uuid.New().String()
	err = job_service.Create(jobID, userID, jobModel.TypeDeleteVM, map[string]interface{}{
		"name": current.Name,
	})

	context.JSONResponse(http.StatusOK, dto.VmDeleted{
		ID:    current.ID,
		JobID: jobID,
	})
}

func isValidSshPublicKey(key string) bool {
	_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
	if err != nil {
		return false
	}
	return true
}
