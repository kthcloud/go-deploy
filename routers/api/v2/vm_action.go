package v2

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/dto/v2/query"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/models/version"
	"github.com/kthcloud/go-deploy/pkg/sys"
	"github.com/kthcloud/go-deploy/service"
	"github.com/kthcloud/go-deploy/service/v2/vms/opts"
)

// CreateVmAction
// @Summary Creates a new action
// @Description Creates a new action
// @Tags VmAction
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param vmId path string true "VM ID"
// @Param body body body.VmActionCreate true "actions body"
// @Success 200 {object} body.VmActionCreated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/vmActions [post]
func CreateVmAction(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.VmActionCreate
	if err := context.GinContext.ShouldBindQuery(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var requestBody body.VmActionCreate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	deployV2 := service.V2(auth)

	vm, err := deployV2.VMs().Get(requestQuery.VmID, opts.GetOpts{Shared: true})
	if err != nil {
		context.ServerError(err, ErrInternal)
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
	err = deployV2.Jobs().Create(jobID, auth.User.ID, model.JobDoVmAction, version.V2, map[string]interface{}{
		"id":       vm.ID,
		"params":   requestBody,
		"authInfo": auth,
	})

	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	context.Ok(body.VmActionCreated{
		ID:    vm.ID,
		JobID: jobID,
	})
}
