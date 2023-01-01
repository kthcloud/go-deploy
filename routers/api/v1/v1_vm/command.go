package v1_vm

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/validator"
	"go-deploy/service/vm_service"
	"net/http"
)

func DoCommand(c *gin.Context) {
	context := app.NewContext(c)

	rules := validator.MapData{
		"vmId": []string{
			"required",
			"uuid_v4",
		},
	}

	bodyRules := validator.MapData{
		"command": []string{
			"required",
			"in:start,stop,reboot",
		},
	}

	validationErrors := context.ValidateParams(&rules)
	if len(validationErrors) > 0 {
		context.ResponseValidationError(validationErrors)
		return
	}

	var commandBody dto.VmCommand
	validationErrors = context.ValidateJSON(&bodyRules, &commandBody)
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

	vm, err := vm_service.GetByID(userID, vmID)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if vm == nil {
		context.NotFound()
		return
	}

	if vm.BeingCreated {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingCreated, "Resource is currently being created")
		return
	}

	if vm.BeingDeleted {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingCreated, "Resource is currently being deleted")
		return
	}

	vm_service.DoCommand(vm, commandBody.Command)

	context.OkDeleted()
}
