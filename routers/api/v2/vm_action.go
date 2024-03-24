package v2

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/dto/v2/body"
	"go-deploy/dto/v2/query"
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	"go-deploy/service/v2/vms/opts"
)

// CreateAction
// @Summary Creates a new action
// @Description Creates a new action
// @Tags VM
// @Accept  json
// @Produce  json
// @Param vmId path string true "VM ID"
// @Param body body body.Action true "Command body"
// @Success 200 {empty} empty
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /vms/{vmId}/command [post]
func CreateAction(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.VmActionCreate
	if err := context.GinContext.ShouldBindQuery(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	var requestBody body.VmActionCreate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)
	deployV2 := service.V2(auth)

	vm, err := deployV2.VMs().Get(requestQuery.VmID, opts.GetOpts{Shared: true})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	if !vm.Ready() {
		context.UserError("VM is not ready")
		return
	}

	jobID := uuid.New().String()
	err = deployV1.Jobs().Create(jobID, auth.UserID, model.JobDoVmAction, version.V2, map[string]interface{}{
		"id":     vm.ID,
		"params": requestBody,
	})

	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(body.VmActionCreated{
		ID:    vm.ID,
		JobID: jobID,
	})
}
