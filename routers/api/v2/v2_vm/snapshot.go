package v2_vm

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto/v2/body"
	"go-deploy/models/dto/v2/query"
	"go-deploy/models/dto/v2/uri"
	"go-deploy/models/sys/job"
	"go-deploy/models/versions"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/v2/utils"
	"go-deploy/service/v2/vms/opts"
)

// ListSnapshots
// @Summary Get snapshot list
// @Description Get snapshot list
// @Tags VM
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param vmId path string true "VM ID"
// @Success 200 {array} body.VmSnapshotRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /vms/{vmId}/snapshots [get]
func ListSnapshots(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.VmSnapshotList
	if err := context.GinContext.ShouldBindQuery(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	var requestURI uri.VmSnapshotList
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	snapshots, err := service.V2().VMs().ListSnapshots(requestURI.VmID, opts.ListSnapshotOpts{
		Pagination: utils.GetOrDefaultPagination(requestQuery.Pagination),
	})
	if err != nil {
		context.ServerError(err, v1.InternalError)
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

	context.Ok(dtoSnapshots)
}

// GetSnapshot
// @Summary Get snapshot
// @Description Get snapshot
// @Tags VM
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param vmId path string true "VM ID"
// @Param snapshotId path string true "Snapshot ID"
// @Success 200 {object} body.VmSnapshotRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /vms/{vmId}/snapshot/{snapshotId} [post]
func GetSnapshot(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.VmSnapshotGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	deployV2 := service.V2(auth)

	vm, err := deployV2.VMs().Get(requestURI.VmID, opts.GetOpts{Shared: true})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	snapshot, err := deployV2.VMs().GetSnapshot(requestURI.VmID, requestURI.SnapshotID, opts.GetSnapshotOpts{})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if snapshot == nil {
		context.NotFound("Snapshot not found")
		return
	}

	context.Ok(snapshot.ToDTOv2())
}

// CreateSnapshot
// @Summary Create snapshot
// @Description Create snapshot
// @Tags VM
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param vmId path string true "VM ID"
// @Success 200 {object} body.VmSnapshotRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /vms/{vmId}/snapshots [post]
func CreateSnapshot(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.VmSnapshotCreate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	var requestBody body.VmSnapshotCreate
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

	err = deployV2.VMs().CheckQuota(requestURI.VmID, auth.UserID, &auth.GetEffectiveRole().Quotas, opts.QuotaOpts{
		CreateSnapshot: &requestBody,
	})
	if err != nil {
		var quotaExceededErr sErrors.QuotaExceededError
		if errors.As(err, &quotaExceededErr) {
			context.Forbidden(quotaExceededErr.Error())
			return
		}

		context.ServerError(err, v1.InternalError)
		return
	}

	current, err := deployV2.VMs().GetSnapshotByName(requestURI.VmID, requestBody.Name, opts.GetSnapshotOpts{})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if current != nil {
		context.UserError("Snapshot already exists")
		return
	}

	vm, err := deployV2.VMs().Get(requestURI.VmID, opts.GetOpts{Shared: true})
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
	err = deployV1.Jobs().Create(jobID, auth.UserID, job.TypeCreateVmUserSnapshot, versions.V2, map[string]interface{}{
		"id": vm.ID,
		"params": body.VmSnapshotCreate{
			Name: requestBody.Name,
		},
	})

	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(body.VmSnapshotCreated{
		ID:    vm.ID,
		JobID: jobID,
	})
}

// DeleteSnapshot
// @Summary Delete snapshot
// @Description Delete snapshot
// @Tags VM
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param vmId path string true "VM ID"
// @Param snapshotId path string true "Snapshot ID"
// @Success 200 {object} body.VmSnapshotRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /vms/{vmId}/snapshot/{snapshotId} [delete]
func DeleteSnapshot(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.VmSnapshotDelete
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
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

	vm, err := deployV2.VMs().Get(requestURI.VmID, opts.GetOpts{Shared: true})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	snapshot, err := deployV2.VMs().GetSnapshot(requestURI.VmID, requestURI.SnapshotID, opts.GetSnapshotOpts{})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if snapshot == nil {
		context.NotFound("Snapshot not found")
		return
	}

	jobID := uuid.New().String()
	err = deployV1.Jobs().Create(jobID, auth.UserID, job.TypeDeleteVmSnapshot, versions.V2, map[string]interface{}{
		"id":         vm.ID,
		"snapshotId": snapshot.ID,
	})

	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(body.VmSnapshotDeleted{
		ID:    vm.ID,
		JobID: jobID,
	})
}
