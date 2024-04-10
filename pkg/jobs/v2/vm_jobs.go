package v2

import (
	"context"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"go-deploy/dto/v2/body"
	"go-deploy/models/model"
	"go-deploy/pkg/db/resources/job_repo"
	"go-deploy/pkg/db/resources/vm_repo"
	jErrors "go-deploy/pkg/jobs/errors"
	"go-deploy/pkg/jobs/utils"
	"go-deploy/pkg/log"
	"go-deploy/pkg/workers/confirm"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/v2/vms/opts"
	"time"
)

func CreateVM(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "ownerId", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	ownerID := job.Args["ownerId"].(string)
	var params body.VmCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2().VMs().Create(id, ownerID, &params)
	if err != nil {
		// TODO: If there was some error, we trigger a repair, since rerunning it would cause a NonUniqueFieldErr
		//_ = service.V2().VMs().Repair(id)
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func DeleteVM(job *model.Job) error {
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

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	go func() {
		err = utils.WaitForJobs(ctx, relatedJobs, []string{model.JobStatusCompleted, model.JobStatusTerminated})
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			if errors.Is(err, context.DeadlineExceeded) {
				log.Println("timeout waiting for related jobs to finish for model", id)
				return
			}

			log.Println("failed to wait for related jobs for model", id, ". details:", err)
		}

		cancel()
	}()

	select {
	case <-time.After(300 * time.Second):
		return jErrors.MakeTerminatedError(fmt.Errorf("timeout waiting for related jobs to finish"))
	case <-ctx.Done():
	}

	err = service.V2().VMs().Delete(id)
	if err != nil {
		if !errors.Is(err, sErrors.VmNotFoundErr) {
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

func UpdateVM(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var update body.VmUpdate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &update)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2().VMs().Update(id, &update)
	if err != nil {
		switch {
		case errors.Is(err, sErrors.VmNotFoundErr):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.NonUniqueFieldErr):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.IngressHostInUseErr):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.NoPortsAvailableErr):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.SnapshotNotFoundErr):
			return jErrors.MakeTerminatedError(err)
		}

		var portInUseErr sErrors.PortInUseErr
		if errors.As(err, &portInUseErr) {
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

func CreateGpuLease(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "userId", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	userID := job.Args["userId"].(string)
	var params body.GpuLeaseCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2(utils.GetAuthInfo(job)).VMs().GpuLeases().Create(id, userID, &params)
	if err != nil {
		switch {
		case errors.Is(err, sErrors.VmNotFoundErr):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.GpuLeaseAlreadyExistsErr):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.GpuNotFoundErr):
			return jErrors.MakeTerminatedError(err)
		}

		return jErrors.MakeFailedError(err)
	}

	return nil
}

func UpdateGpuLease(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var params body.GpuLeaseUpdate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2().VMs().GpuLeases().Update(id, &params)
	if err != nil {
		switch {
		case errors.Is(err, sErrors.GpuLeaseNotFoundErr):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.GpuLeaseNotAssignedErr):
			return jErrors.MakeTerminatedError(err)
		}

		return jErrors.MakeFailedError(err)
	}

	return nil
}

func DeleteGpuLease(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = service.V2().VMs().GpuLeases().Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, sErrors.GpuNotFoundErr):
			return jErrors.MakeTerminatedError(err)
		}

		return jErrors.MakeFailedError(err)
	}

	return nil
}

func CreateSystemVmSnapshot(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)
	var params model.CreateSnapshotParams
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	_, err = service.V2().VMs().Snapshots().Create(vmID, opts.CreateSnapshotOpts{System: &params})
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func CreateUserVmSnapshot(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)
	var params body.VmSnapshotCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	_, err = service.V2().VMs().Snapshots().Create(vmID, opts.CreateSnapshotOpts{User: &params})
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func DeleteVmSnapshot(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "snapshotId"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)
	snapshotID := job.Args["snapshotId"].(string)

	err = service.V2().VMs().Snapshots().Delete(vmID, snapshotID)
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func DoVmAction(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)
	var params body.VmActionCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2().VMs().DoAction(vmID, &params)
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}
