package v1_vm

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/uri"
	"go-deploy/models/sys/job"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/job_service"
	"go-deploy/service/vm_service"
	"net/http"
)

// GetSnapshotList
// @Summary Get snapshot list
// @Description Get snapshot list
// @Tags VM
// @Accept json
// @Produce json
// @Param Authorization header string true "With the bearer started"
// @Param vmId path string true "VM ID"
// @Success 200 {array} body.VmSnapshotRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
func GetSnapshotList(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.VmCommand
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	snapshots, _ := vm_service.GetSnapshotsByVM(requestURI.VmID)
	if snapshots == nil {
		context.JSONResponse(200, []interface{}{})
		return
	}

	dtoSnapshots := make([]body.VmSnapshotRead, len(snapshots))
	for i, snapshot := range snapshots {
		dtoSnapshots[i] = body.VmSnapshotRead{
			ID:         snapshot.ID,
			VmID:       snapshot.VmID,
			Name:       snapshot.Name,
			ParentName: snapshot.ParentName,
			CreatedAt:  snapshot.CreatedAt,
			State:      snapshot.State,
			Current:    snapshot.Current,
		}
	}

	context.JSONResponse(200, dtoSnapshots)
}

func CreateSnapshot(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.VmSnapshotCreate
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	var requestBody body.VmSnapshotCreate
	if err := context.GinContext.BindJSON(&requestBody); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	ok, reason, err := vm_service.CheckQuotaCreateSnapshot(auth.UserID, &auth.GetEffectiveRole().Quotas)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to check quota: %s", err))
		return
	}

	if !ok {
		context.ErrorResponse(http.StatusForbidden, status_codes.Error, fmt.Sprintf("Failed to create snapshot: %s", reason))
		return
	}

	current, err := vm_service.GetSnapshotByName(requestURI.VmID, requestBody.Name)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to check if a snapshot with same name exists: %s", err))
		return
	}

	if current != nil {
		context.ErrorResponse(http.StatusConflict, status_codes.Error, "Snapshot already exists with given name")
		return
	}

	vm, err := vm_service.GetByID(auth.UserID, requestURI.VmID, auth.IsAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get vm: %s", err))
		return
	}

	if vm == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.Error, fmt.Sprintf("VM %s not found", requestURI.VmID))
		return
	}

	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, job.TypeCreateSnapshot, map[string]interface{}{
		"id":          vm.ID,
		"name":        requestBody.Name,
		"userCreated": true,
	})

	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to create job: %s", err))
		return
	}

	context.JSONResponse(http.StatusCreated, body.VmCreated{
		ID:    vm.ID,
		JobID: jobID,
	})
}
