package v1_vm

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/uri"
	"go-deploy/models/sys/job"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	"go-deploy/service/job_service"
	"go-deploy/service/vm_service"
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

	var requestURI uri.VmSnapshotList
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	snapshots, _ := vm_service.ListSnapshotsByVM(requestURI.VmID)
	if snapshots == nil {
		context.Ok([]interface{}{})
		return
	}

	dtoSnapshots := make([]body.VmSnapshotRead, len(snapshots))
	for i, snapshot := range snapshots {
		dtoSnapshots[i] = snapshot.ToDTO()
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

	vm, err := vm_service.GetByIdAuth(requestURI.VmID, auth)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	snapshot, err := vm_service.GetSnapshot(requestURI.VmID, requestURI.SnapshotID)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if snapshot == nil {
		context.NotFound("Snapshot not found")
		return
	}

	context.Ok(snapshot.ToDTO())
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

	err = vm_service.CheckQuotaCreateSnapshot(auth.UserID, &auth.GetEffectiveRole().Quotas, auth)
	if err != nil {
		var quotaExceededErr service.QuotaExceededError
		if errors.As(err, &quotaExceededErr) {
			context.Forbidden(quotaExceededErr.Error())
			return
		}

		context.ServerError(err, v1.InternalError)
		return
	}

	current, err := vm_service.ListSnapshotsByName(requestURI.VmID, requestBody.Name)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if current != nil {
		context.UserError("Snapshot already exists")
		return
	}

	vm, err := vm_service.GetByIdAuth(requestURI.VmID, auth)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, job.TypeCreateUserSnapshot, map[string]interface{}{
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

	vm, err := vm_service.GetByIdAuth(requestURI.VmID, auth)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	snapshot, err := vm_service.GetSnapshot(requestURI.VmID, requestURI.SnapshotID)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if snapshot == nil {
		context.NotFound("Snapshot not found")
		return
	}

	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, job.TypeDeleteSnapshot, map[string]interface{}{
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
