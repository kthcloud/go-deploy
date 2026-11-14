package v2

import (
	"context"
	"errors"
	"fmt"
	"time"

	body "github.com/kthcloud/go-deploy/api/openapi/v2"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/job_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/vm_repo"
	jErrors "github.com/kthcloud/go-deploy/pkg/jobs/errors"
	"github.com/kthcloud/go-deploy/pkg/jobs/utils"
	"github.com/kthcloud/go-deploy/pkg/services/confirm"
	"github.com/kthcloud/go-deploy/service"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	"github.com/kthcloud/go-deploy/service/v2/vms/opts"
	"github.com/mitchellh/mapstructure"
)

func CreateVM(ctx context.Context, job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "ownerId", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	ownerID := job.Args["ownerId"].(string)
	var params body.VmCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]any), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2(utils.GetAuthInfo(job)).VMs().Create(ctx, id, ownerID, &params)
	if err != nil {
		// If there was some error, we trigger a repair, since rerunning it would cause a ErrNonUniqueField
		_ = service.V2(utils.GetAuthInfo(job)).VMs().Repair(ctx, id)
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func DeleteVM(ctx context.Context, job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = vm_repo.New().AddActivity(id, model.ActivityBeingDeleted)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	relatedJobs, err := job_repo.New().
		ExcludeScheduled().
		ExcludeTypes(model.JobDeleteVM).
		ExcludeStatus(model.JobStatusTerminated, model.JobStatusCompleted).
		ExcludeIDs(job.ID).
		FilterArgs("id", id).
		List()
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	if err := utils.WaitForRelatedJobs(ctx, id, relatedJobs, 300*time.Second); err != nil {
		if errors.Is(err, context.Canceled) {
			return jErrors.MakeFailedError(err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return jErrors.MakeTerminatedError(fmt.Errorf("timeout waiting for related jobs to finish: %w", err))
		}
		return jErrors.MakeFailedError(err)
	}

	err = service.V2(utils.GetAuthInfo(job)).VMs().Delete(ctx, id)
	if err != nil {
		if !errors.Is(err, sErrors.ErrVmNotFound) {
			return jErrors.MakeFailedError(err)
		}
	}

	// Check if deleted, otherwise mark as failed and return to queue for retry
	vm, err := vm_repo.New().GetByID(id)
	if err != nil {
		return jErrors.MakeFailedError(err)
	}

	if vm != nil {
		if confirm.VmDeleted(vm) {
			return nil
		}

		return jErrors.MakeFailedError(fmt.Errorf("vm not deleted"))
	}

	return nil
}

func UpdateVM(ctx context.Context, job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var update body.VmUpdate
	err = mapstructure.Decode(job.Args["params"].(map[string]any), &update)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2(utils.GetAuthInfo(job)).VMs().Update(ctx, id, &update)
	if err != nil {
		switch {
		case errors.Is(err, sErrors.ErrVmNotFound):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.ErrNonUniqueField):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.ErrIngressHostInUse):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.ErrNoPortsAvailable):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.ErrSnapshotNotFound):
			return jErrors.MakeTerminatedError(err)
		}

		return jErrors.MakeFailedError(err)
	}

	err = vm_repo.New().MarkUpdated(id)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func CreateGpuLease(ctx context.Context, job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "userId", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	userID := job.Args["userId"].(string)
	var params body.GpuLeaseCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]any), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2(utils.GetAuthInfo(job)).VMs().GpuLeases().Create(ctx, id, userID, &params)
	if err != nil {
		switch {
		case errors.Is(err, sErrors.ErrVmNotFound):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.ErrGpuLeaseAlreadyExists):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.ErrGpuNotFound):
			return jErrors.MakeTerminatedError(err)
		}

		return jErrors.MakeFailedError(err)
	}

	return nil
}

func UpdateGpuLease(ctx context.Context, job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var params body.GpuLeaseUpdate
	err = mapstructure.Decode(job.Args["params"].(map[string]any), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2(utils.GetAuthInfo(job)).VMs().GpuLeases().Update(ctx, id, &params)
	if err != nil {
		switch {
		case errors.Is(err, sErrors.ErrGpuLeaseNotFound):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.ErrGpuLeaseNotAssigned):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.ErrVmAlreadyAttached):
			return jErrors.MakeTerminatedError(err)
		}

		return jErrors.MakeFailedError(err)
	}

	return nil
}

func DeleteGpuLease(ctx context.Context, job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = service.V2(utils.GetAuthInfo(job)).VMs().GpuLeases().Delete(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, sErrors.ErrGpuNotFound):
			return jErrors.MakeTerminatedError(err)
		}

		return jErrors.MakeFailedError(err)
	}

	return nil
}

func CreateSystemVmSnapshot(ctx context.Context, job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)
	var params model.CreateSnapshotParams
	err = mapstructure.Decode(job.Args["params"].(map[string]any), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	_, err = service.V2(utils.GetAuthInfo(job)).VMs().Snapshots().Create(ctx, vmID, opts.CreateSnapshotOpts{System: &params})
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func CreateUserVmSnapshot(ctx context.Context, job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)
	var params body.VmSnapshotCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]any), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	_, err = service.V2(utils.GetAuthInfo(job)).VMs().Snapshots().Create(ctx, vmID, opts.CreateSnapshotOpts{User: &params})
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func DeleteVmSnapshot(ctx context.Context, job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "snapshotId"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)
	snapshotID := job.Args["snapshotId"].(string)

	err = service.V2(utils.GetAuthInfo(job)).VMs().Snapshots().Delete(ctx, vmID, snapshotID)
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func DoVmAction(ctx context.Context, job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)
	var params body.VmActionCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]any), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2(utils.GetAuthInfo(job)).VMs().DoAction(ctx, vmID, &params)
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func UpdateVmOwner(ctx context.Context, job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var params model.VmUpdateOwnerParams
	err = mapstructure.Decode(job.Args["params"].(map[string]any), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2(utils.GetAuthInfo(job)).VMs().UpdateOwner(ctx, id, &params)
	if err != nil {
		if errors.Is(err, sErrors.ErrVmNotFound) {
			return jErrors.MakeTerminatedError(err)
		}

		return jErrors.MakeFailedError(err)
	}

	if job.HasArg("resourceMigrationId") {
		resourceMigrationID := job.Args["resourceMigrationId"].(string)
		err = service.V2().ResourceMigrations().Delete(ctx, resourceMigrationID)
		if err != nil {
			return jErrors.MakeTerminatedError(err)
		}
	}

	err = vm_repo.New().MarkUpdated(id)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func RepairVM(ctx context.Context, job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = service.V2(utils.GetAuthInfo(job)).VMs().Repair(ctx, id)
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return jErrors.MakeTerminatedError(err)
	}

	err = vm_repo.New().MarkRepaired(id)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}
