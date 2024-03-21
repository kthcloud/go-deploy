package v1

import (
	"context"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"go-deploy/dto/v1/body"
	"go-deploy/models/model"
	"go-deploy/pkg/db/resources/deployment_repo"
	"go-deploy/pkg/db/resources/job_repo"
	"go-deploy/pkg/db/resources/vm_repo"
	jErrors "go-deploy/pkg/jobs/errors"
	"go-deploy/pkg/jobs/utils"
	"go-deploy/pkg/workers/confirm"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/v1/vms/opts"
	"log"
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

	err = service.V1().VMs().Create(id, ownerID, &params)
	if err != nil {
		// If there was some error, we trigger a repair, since rerunning it would cause a NonUniqueFieldErr
		_ = service.V1().VMs().Repair(id)
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

	err = service.V1().VMs().Delete(id)
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

	err = service.V1().VMs().Update(id, &update)
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

func UpdateVmOwner(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var params body.VmUpdateOwner
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V1().VMs().UpdateOwner(id, &params)
	if err != nil {
		return jErrors.MakeFailedError(err)
	}

	return nil
}

func AttachGPU(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "gpuIds", "userId", "leaseDuration"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)
	var gpuIDs []string
	err = mapstructure.Decode(job.Args["gpuIds"].(interface{}), &gpuIDs)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}
	leaseDuration := job.Args["leaseDuration"].(float64)

	// We keep this field to know who requested the gpu_repo attachment
	_ = job.Args["userId"].(string)

	err = service.V1().VMs().AttachGPU(vmID, gpuIDs, leaseDuration)
	if err != nil {
		if errors.Is(err, sErrors.GpuNotFoundErr) {
			return jErrors.MakeTerminatedError(err)
		}

		if errors.Is(err, sErrors.VmNotFoundErr) {
			return jErrors.MakeTerminatedError(err)
		}

		return jErrors.MakeFailedError(err)
	}

	return nil
}

func DetachGpuFromVM(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)

	err = service.V1().VMs().DetachGPU(vmID)
	if err != nil {
		return jErrors.MakeFailedError(err)
	}

	return nil
}

func CreateDeployment(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "ownerId", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	ownerID := job.Args["ownerId"].(string)
	var params body.DeploymentCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V1().Deployments().Create(id, ownerID, &params)
	if err != nil {
		// If there was some error, we trigger a repair, since rerunning it would cause a NonUniqueFieldErr
		_ = service.V1().Deployments().Repair(id)
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func DeleteDeployment(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = deployment_repo.New().AddActivity(id, model.ActivityBeingDeleted)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	relatedJobs, err := job_repo.New().
		ExcludeScheduled().
		ExcludeTypes(model.JobDeleteDeployment).
		ExcludeStatus(model.JobStatusTerminated, model.JobStatusCompleted).
		ExcludeIDs(job.ID).
		FilterArgs("id", id).
		List()
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	case <-time.After(30 * time.Second):
		return jErrors.MakeTerminatedError(fmt.Errorf("timeout waiting for related jobs to finish"))
	case <-ctx.Done():
	}

	err = service.V1().Deployments().Delete(id)
	if err != nil {
		if !errors.Is(err, sErrors.DeploymentNotFoundErr) {
			return jErrors.MakeFailedError(err)
		}
	}

	// Check if deleted, otherwise mark as failed and return to queue for retry
	deployment, err := deployment_repo.New().GetByID(id)
	if err != nil {
		return jErrors.MakeFailedError(err)
	}

	if deployment != nil {
		if confirm.DeploymentDeleted(deployment) {
			return nil
		}

		return jErrors.MakeFailedError(fmt.Errorf("deployment not deleted"))
	}

	return nil
}

func UpdateDeployment(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var update body.DeploymentUpdate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &update)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V1().Deployments().Update(id, &update)
	if err != nil {
		switch {
		case errors.Is(err, sErrors.DeploymentNotFoundErr):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.NonUniqueFieldErr):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.IngressHostInUseErr):
			return jErrors.MakeTerminatedError(err)
		}

		return jErrors.MakeFailedError(err)
	}

	err = deployment_repo.New().MarkUpdated(id)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func UpdateDeploymentOwner(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var params body.DeploymentUpdateOwner
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V1().Deployments().UpdateOwner(id, &params)
	if err != nil {
		if errors.Is(err, sErrors.DeploymentNotFoundErr) {
			return jErrors.MakeTerminatedError(err)
		}

		return jErrors.MakeFailedError(err)
	}

	return nil
}

func RepairDeployment(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = service.V1().Deployments().Repair(id)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func CreateSM(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "userId", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	userID := job.Args["userId"].(string)

	var params model.SmCreateParams
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V1().SMs().Create(id, userID, &params)
	if err != nil {
		// We always terminate these jobs, since rerunning it would cause a NonUniqueFieldErr
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func DeleteSM(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = service.V1().SMs().Delete(id)
	if err != nil {
		return jErrors.MakeFailedError(err)
	}

	return nil
}

func RepairSM(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = service.V1().SMs().Repair(id)
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func RepairVM(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = service.V1().VMs().Repair(id)
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func CreateSystemSnapshot(job *model.Job) error {
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

	err = service.V1().VMs().CreateSnapshot(vmID, &opts.CreateSnapshotOpts{System: &params})
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func CreateUserSnapshot(job *model.Job) error {
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

	err = service.V1().VMs().CreateSnapshot(vmID, &opts.CreateSnapshotOpts{User: &params})
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func DeleteSnapshot(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "snapshotId"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)
	snapshotID := job.Args["snapshotId"].(string)

	err = service.V1().VMs().DeleteSnapshot(vmID, snapshotID)
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}
