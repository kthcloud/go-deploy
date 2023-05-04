package v1_vm

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/uri"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/vm_service"
	"net/http"
)

// DoCommand
// @Summary Do command
// @Description Do command
// @Tags VM
// @Accept  json
// @Produce  json
// @Param vmId path string true "VM ID"
// @Param body body body.DoCommand true "Command body"
// @Success 200 {empty} empty
// @Failure 400 {object} app.ErrorResponse
// @Failure 404 {object} app.ErrorResponse
// @Failure 423 {object} app.ErrorResponse
// @Failure 500 {object} app.ErrorResponse
// @Router /api/v1/vms/{vmId}/command [post]
func DoCommand(c *gin.Context) {
	context := app.NewContext(c)

	var requestURI uri.DoCommand
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	var requestBody body.DoCommand
	if err := context.GinContext.BindJSON(&requestBody); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	vm, err := vm_service.GetByID(auth.UserID, requestURI.VmID, auth.IsAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if vm == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("VM with id %s not found", requestURI.VmID))
		return
	}

	if vm.BeingCreated {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingCreated, "Resource is currently being created")
		return
	}

	if vm.BeingDeleted {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingDeleted, "Resource is currently being deleted")
		return
	}

	vm_service.DoCommand(vm, requestBody.Command)

	context.OkDeleted()
}
