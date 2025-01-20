package v2

import (
	"github.com/gin-gonic/gin"
	"github.com/kthcloud/go-deploy/pkg/sys"
)

// GetSnapshot
// @Summary Get snapshot
// @Description Get snapshot
// @Tags Snapshot
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param vmId path string true "VM ID"
// @Param snapshotId path string true "Snapshot ID"
// @Success 200 {object} body.VmSnapshotRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/vms/{vmId}/snapshot/{snapshotId} [post]
func GetSnapshot(c *gin.Context) {
	context := sys.NewContext(c)

	// Not yet implemented
	context.NotImplemented()

	/*var requestURI uri.VmSnapshotGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	deployV2 := service.V2(auth)

	vm, err := deployV2.VMs().Get(requestURI.VmID, opts.GetOpts{Shared: true})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	snapshot, err := deployV2.VMs().Snapshots().Get(requestURI.VmID, requestURI.SnapshotID, opts.GetSnapshotOpts{})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if snapshot == nil {
		context.NotFound("Snapshot not found")
		return
	}

	context.Ok(snapshot.ToDTOv2())*/
}

// ListSnapshots
// @Summary List snapshots
// @Description List snapshots
// @Tags Snapshot
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param vmId path string true "Filter by VM ID"
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {array} body.VmSnapshotRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/snapshots [get]
func ListSnapshots(c *gin.Context) {
	context := sys.NewContext(c)

	// Not yet implemented
	context.NotImplemented()

	/*var requestQuery query.VmSnapshotList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var requestURI uri.VmSnapshotList
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	snapshots, err := service.V2().VMs().Snapshots().List(requestURI.VmID, opts.ListSnapshotOpts{
		Pagination: utils.GetOrDefaultPagination(requestQuery.Pagination),
	})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if snapshots == nil {
		context.Ok([]interface{}{})
		return
	}

	dtoSnapshots := make([]body.VmSnapshotRead, len(snapshots))
	for i, snapshot := range snapshots {
		dtoSnapshots[i] = snapshot.ToDTOv2()
	}

	context.Ok(dtoSnapshots)*/
}

// CreateSnapshot
// @Summary Create snapshot
// @Description Create snapshot
// @Tags Snapshot
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param vmId path string true "VM ID"
// @Success 200 {object} body.VmSnapshotCreated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/snapshots [post]
func CreateSnapshot(c *gin.Context) {
	context := sys.NewContext(c)

	// Not yet implemented
	context.NotImplemented()

	/*var requestURI uri.VmSnapshotCreate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var requestBody body.VmSnapshotCreate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	deployV2 := service.V2(auth)

	err = deployV2.VMs().CheckQuota(requestURI.VmID, auth.User.ID, &auth.GetEffectiveRole().Quotas, opts.QuotaOpts{
		CreateSnapshot: &requestBody,
	})
	if err != nil {
		var quotaExceededErr sErrors.QuotaExceededError
		if errors.As(err, &quotaExceededErr) {
			context.Forbidden(quotaExceededErr.Error())
			return
		}

		context.ServerError(err, InternalError)
		return
	}

	current, err := deployV2.VMs().Snapshots().GetByName(requestURI.VmID, requestBody.Name, opts.GetSnapshotOpts{})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if current != nil {
		context.UserError("Snapshot already exists")
		return
	}

	vm, err := deployV2.VMs().Get(requestURI.VmID, opts.GetOpts{Shared: true})
	if err != nil {
		context.ServerError(err, InternalError)
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
	err = deployV2.Jobs().Create(jobID, auth.User.ID, model.JobCreateVmUserSnapshot, version.V2, map[string]interface{}{
		"id": vm.ID,
		"params": body.VmSnapshotCreate{
			Name: requestBody.Name,
		},
		"authInfo": auth,
	})

	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	context.Ok(body.VmSnapshotCreated{
		ID:    vm.ID,
		JobID: jobID,
	})*/
}

// DeleteSnapshot
// @Summary Delete snapshot
// @Description Delete snapshot
// @Tags Snapshot
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param vmId path string true "VM ID"
// @Param snapshotId path string true "Snapshot ID"
// @Success 200 {object} body.VmSnapshotDeleted
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/vms/{vmId}/snapshot/{snapshotId} [delete]
func DeleteSnapshot(c *gin.Context) {
	context := sys.NewContext(c)

	// Not yet implemented
	context.NotImplemented()

	/*var requestURI uri.VmSnapshotDelete
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	deployV2 := service.V2(auth)

	vm, err := deployV2.VMs().Get(requestURI.VmID, opts.GetOpts{Shared: true})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	snapshot, err := deployV2.VMs().Snapshots().Get(requestURI.VmID, requestURI.SnapshotID, opts.GetSnapshotOpts{})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if snapshot == nil {
		context.NotFound("Snapshot not found")
		return
	}

	jobID := uuid.New().String()
	err = deployV2.Jobs().Create(jobID, auth.User.ID, model.JobDeleteVmSnapshot, version.V2, map[string]interface{}{
		"id":         vm.ID,
		"snapshotId": snapshot.ID,
		"authInfo":   auth,
	})

	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	context.Ok(body.VmSnapshotDeleted{
		ID:    vm.ID,
		JobID: jobID,
	})*/
}
