package v1_vm

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/validator"
	"go-deploy/service/vm_service"
	"net/http"
	"strconv"
)

func getAllVMs(userID string, context *app.ClientContext) {
	vms, _ := vm_service.GetAll()

	dtoVMs := make([]dto.VmRead, len(vms))
	for i, vm := range vms {
		_, statusMsg, _ := vm_service.GetStatusByID(userID, vm.ID)
		dtoVMs[i] = vm.ToDto(statusMsg)
	}

	context.JSONResponse(http.StatusOK, dtoVMs)
}

func GetMany(c *gin.Context) {
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

	// might want to check if userID is allowed to get all...
	wantAll, _ := strconv.ParseBool(context.GinContext.Query("all"))
	if wantAll {
		getAllVMs(userID, &context)
		return
	}

	vms, _ := vm_service.GetByOwnerID(userID)
	if vms == nil {
		context.JSONResponse(200, []interface{}{})
		return
	}

	dtoVMs := make([]dto.VmRead, len(vms))
	for i, vm := range vms {
		_, statusMsg, _ := vm_service.GetStatusByID(userID, vm.ID)
		dtoVMs[i] = vm.ToDto(statusMsg)
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

	vm, _ := vm_service.GetByID(userID, vmID)

	if vm == nil {
		context.NotFound()
		return
	}

	_, statusMsg, _ := vm_service.GetStatusByID(userID, vm.ID)
	context.JSONResponse(200, vm.ToDto(statusMsg))
}

func Create(c *gin.Context) {
	context := app.NewContext(c)

	bodyRules := validator.MapData{
		"name": []string{
			"required",
			"regex:^[a-zA-Z]+$",
			"min:3",
			"max:10",
		},
	}

	messages := validator.MapData{
		"name": []string{
			"required:Name is required",
			"regexp:Name must be all lowercase",
			"min:Name must be between 3-10 characters",
			"max:Name must be between 3-10 characters",
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
	userId := token.Sub

	exists, vm, err := vm_service.Exists(requestBody.Name)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if exists {
		if vm.Owner != userId {
			context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceAlreadyExists, "Resource already exists")
			return
		}
		if vm.BeingDeleted {
			context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingDeleted, "Resource is currently being deleted")
			return
		}
		vm_service.Create(vm.ID, requestBody.Name, userId)
		context.JSONResponse(http.StatusCreated, dto.VmCreated{ID: vm.ID})
		return
	}

	vmID := uuid.New().String()
	vm_service.Create(vmID, requestBody.Name, userId)
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
	userId := token.Sub
	vmID := context.GinContext.Param("vmId")

	current, err := vm_service.GetByID(userId, vmID)
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
