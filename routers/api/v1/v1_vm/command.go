package v1_vm

import (
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/uri"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/vm_service"
	"go-deploy/service/vm_service/client"
)

// DoCommand
// @Summary Do command
// @Description Do command
// @Tags VM
// @Accept  json
// @Produce  json
// @Param vmId path string true "VM ID"
// @Param body body body.VmCommand true "Command body"
// @Success 200 {empty} empty
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /vms/{vmId}/command [post]
func DoCommand(c *gin.Context) {
	// TODO: this route is weird and should be covered in the vm update route with desired states

	context := sys.NewContext(c)

	var requestURI uri.VmCommand
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	var requestBody body.VmCommand
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	vsc := vm_service.New().WithAuth(auth)

	vm, err := vsc.Get(requestURI.VmID, client.GetOptions{Shared: true})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	if !vm.Ready() {
		context.Locked("VM is not ready")
		return
	}

	vsc.DoCommand(requestURI.VmID, requestBody.Command)

	context.OkNoContent()
}
