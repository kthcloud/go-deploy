package v2

import (
	"context"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"go-deploy/models/dto/v2/body"
	jobModels "go-deploy/models/sys/job"
	vmModels "go-deploy/models/sys/vm"
	errors2 "go-deploy/pkg/jobs/errors"
	"go-deploy/pkg/jobs/utils"
	"go-deploy/pkg/workers/confirm"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/v2/vms/opts"
	"log"
	"time"
)

func CreateVM(job *jobModels.Job) error {
	err := utils.AssertParameters(job, []string{"id", "ownerId", "params"})
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	ownerID := job.Args["ownerId"].(string)
	var params body.VmCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	err = service.V2().VMs().Create(id, ownerID, &params)
	if err != nil {
		// TODO: If there was some error, we trigger a repair, since rerunning it would cause a NonUniqueFieldErr
		//_ = service.V2().VMs().Repair(id)
		return errors2.MakeTerminatedError(err)
	}

	return nil
}

func DeleteVM(job *jobModels.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = vmModels.New().AddActivity(id, vmModels.ActivityBeingDeleted)
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	relatedJobs, err := jobModels.New().
		ExcludeScheduled().
		ExcludeTypes(jobModels.TypeDeleteVM).
		ExcludeStatus(jobModels.StatusTerminated, jobModels.StatusCompleted).
		ExcludeIDs(job.ID).
		FilterArgs("id", id).
		List()
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	go func() {
		err = utils.WaitForJobs(ctx, relatedJobs, []string{jobModels.StatusCompleted, jobModels.StatusTerminated})
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			if errors.Is(err, context.DeadlineExceeded) {
				log.Println("timeout waiting for related jobs to finish for resource", id)
				return
			}

			log.Println("failed to wait for related jobs for resource", id, ". details:", err)
		}

		cancel()
	}()

	select {
	case <-time.After(300 * time.Second):
		return errors2.MakeTerminatedError(fmt.Errorf("timeout waiting for related jobs to finish"))
	case <-ctx.Done():
	}

	err = service.V2().VMs().Delete(id)
	if err != nil {
		if !errors.Is(err, sErrors.VmNotFoundErr) {
			return errors2.MakeFailedError(err)
		}
	}

	// Check if deleted, otherwise mark as failed and return to queue for retry
	vm, err := vmModels.New().GetByID(id)
	if err != nil {
		return errors2.MakeFailedError(err)
	}

	if vm != nil {
		if confirm.VmDeleted(vm) {
			return nil
		}

		return errors2.MakeFailedError(fmt.Errorf("vm not deleted"))
	}

	return nil
}

func UpdateVM(job *jobModels.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var update body.VmUpdate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &update)
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	err = service.V2().VMs().Update(id, &update)
	if err != nil {
		switch {
		case errors.Is(err, sErrors.VmNotFoundErr):
			return errors2.MakeTerminatedError(err)
		case errors.Is(err, sErrors.NonUniqueFieldErr):
			return errors2.MakeTerminatedError(err)
		case errors.Is(err, sErrors.IngressHostInUseErr):
			return errors2.MakeTerminatedError(err)
		case errors.Is(err, sErrors.NoPortsAvailableErr):
			return errors2.MakeTerminatedError(err)
		case errors.Is(err, sErrors.SnapshotNotFoundErr):
			return errors2.MakeTerminatedError(err)
		}

		var portInUseErr sErrors.PortInUseErr
		if errors.As(err, &portInUseErr) {
			return errors2.MakeTerminatedError(err)
		}

		return errors2.MakeFailedError(err)
	}

	err = vmModels.New().MarkUpdated(id)
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	return nil
}

func CreateSystemSnapshot(job *jobModels.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)
	var params vmModels.CreateSnapshotParams
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	_, err = service.V2().VMs().CreateSnapshot(vmID, &opts.CreateSnapshotOpts{System: &params})
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return errors2.MakeTerminatedError(err)
	}

	return nil
}

func CreateUserSnapshot(job *jobModels.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)
	var params body.VmSnapshotCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	_, err = service.V2().VMs().CreateSnapshot(vmID, &opts.CreateSnapshotOpts{User: &params})
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return errors2.MakeTerminatedError(err)
	}

	return nil
}

func DeleteSnapshot(job *jobModels.Job) error {
	err := utils.AssertParameters(job, []string{"id", "snapshotId"})
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)
	snapshotID := job.Args["snapshotId"].(string)

	err = service.V2().VMs().DeleteSnapshot(vmID, snapshotID)
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return errors2.MakeTerminatedError(err)
	}

	return nil
}
