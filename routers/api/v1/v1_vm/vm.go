package v1_vm

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/validator"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/user_info_service"
	"go-deploy/service/vm_service"
	"golang.org/x/crypto/ssh"
	"net/http"
	"strconv"
)

func getAllVMs(context *app.ClientContext) {
	vms, _ := vm_service.GetAll()

	dtoVMs := make([]dto.VmRead, len(vms))
	for i, vm := range vms {
		connectionString, _ := vm_service.GetConnectionString(&vm)
		dtoVMs[i] = vm.ToDto(vm.StatusMessage, connectionString)
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
		dtoVMs[i] = vm.ToDto(vm.StatusMessage, connectionString)
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

	vm, _ := vm_service.GetByID(userID, vmID, isAdmin)

	if vm == nil {
		context.NotFound()
		return
	}

	connectionString, _ := vm_service.GetConnectionString(vm)
	context.JSONResponse(200, vm.ToDto(vm.StatusMessage, connectionString))
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
		context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceValidationFailed, "SSH public key is invalid")
		return
	}

	exists, vm, err := vm_service.Exists(requestBody.Name)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if exists {
		if vm.OwnerID != userID {
			context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceAlreadyExists, "Resource already exists")
			return
		}
		if vm.BeingDeleted {
			context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingDeleted, "Resource is currently being deleted")
			return
		}
		vm_service.Create(vm.ID, requestBody.Name, requestBody.SshPublicKey, userID)
		context.JSONResponse(http.StatusCreated, dto.VmCreated{ID: vm.ID})
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
	vm_service.Create(vmID, requestBody.Name, requestBody.SshPublicKey, userID)
	context.JSONResponse(http.StatusCreated, dto.VmCreated{ID: vmID})
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
		context.NotFound()
		return
	}

	if current.BeingCreated {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingCreated, "Resource is currently being created")
		return
	}

	if !current.BeingDeleted {
		_ = vm_service.MarkBeingDeleted(current.ID)
	}

	vm_service.Delete(current.Name)

	context.OkDeleted()
}

func isValidSshPublicKey(key string) bool {
	_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
	if err != nil {
		return false
	}
	return true
}
