package v1

import (
	"context"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"go-deploy/models/dto/v1/body"
	deploymentModels "go-deploy/models/sys/deployment"
	jobModels "go-deploy/models/sys/job"
	smModels "go-deploy/models/sys/sm"
	vmModels "go-deploy/models/sys/vm"
	errors2 "go-deploy/pkg/jobs/errors"
	"go-deploy/pkg/jobs/utils"
	"go-deploy/pkg/workers/confirm"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/v1/vms/opts"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	err = service.V1().VMs().Create(id, ownerID, &params)
	if err != nil {
		// If there was some error, we trigger a repair, since rerunning it would cause a NonUniqueFieldErr
		_ = service.V1().VMs().Repair(id)
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

	err = service.V1().VMs().Delete(id)
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

	err = service.V1().VMs().Update(id, &update)
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

func UpdateVmOwner(job *jobModels.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var params body.VmUpdateOwner
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	err = service.V1().VMs().UpdateOwner(id, &params)
	if err != nil {
		return errors2.MakeFailedError(err)
	}

	return nil
}

func AttachGpuToVM(job *jobModels.Job) error {
	err := utils.AssertParameters(job, []string{"id", "gpuIds", "userId", "leaseDuration"})
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)
	var gpuIDs []string
	err = mapstructure.Decode(job.Args["gpuIds"].(interface{}), &gpuIDs)
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}
	leaseDuration := job.Args["leaseDuration"].(float64)

	// We keep this field to know who requested the gpu attachment
	_ = job.Args["userId"].(string)

	err = service.V1().VMs().AttachGPU(vmID, gpuIDs, leaseDuration)
	if err != nil {
		if errors.Is(err, sErrors.GpuNotFoundErr) {
			return errors2.MakeTerminatedError(err)
		}

		if errors.Is(err, sErrors.VmNotFoundErr) {
			return errors2.MakeTerminatedError(err)
		}

		return errors2.MakeFailedError(err)
	}

	return nil
}

func DetachGpuFromVM(job *jobModels.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)

	err = service.V1().VMs().DetachGPU(vmID)
	if err != nil {
		return errors2.MakeFailedError(err)
	}

	return nil
}

func CreateDeployment(job *jobModels.Job) error {
	err := utils.AssertParameters(job, []string{"id", "ownerId", "params"})
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	ownerID := job.Args["ownerId"].(string)
	var params body.DeploymentCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	err = service.V1().Deployments().Create(id, ownerID, &params)
	if err != nil {
		// If there was some error, we trigger a repair, since rerunning it would cause a NonUniqueFieldErr
		_ = service.V1().Deployments().Repair(id)
		return errors2.MakeTerminatedError(err)
	}

	return nil
}

func DeleteDeployment(job *jobModels.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = deploymentModels.New().AddActivity(id, deploymentModels.ActivityBeingDeleted)
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	relatedJobs, err := jobModels.New().
		ExcludeScheduled().
		ExcludeTypes(jobModels.TypeDeleteDeployment).
		ExcludeStatus(jobModels.StatusTerminated, jobModels.StatusCompleted).
		ExcludeIDs(job.ID).
		FilterArgs("id", id).
		List()
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	case <-time.After(30 * time.Second):
		return errors2.MakeTerminatedError(fmt.Errorf("timeout waiting for related jobs to finish"))
	case <-ctx.Done():
	}

	err = service.V1().Deployments().Delete(id)
	if err != nil {
		if !errors.Is(err, sErrors.DeploymentNotFoundErr) {
			return errors2.MakeFailedError(err)
		}
	}

	// Check if deleted, otherwise mark as failed and return to queue for retry
	deployment, err := deploymentModels.New().GetByID(id)
	if err != nil {
		return errors2.MakeFailedError(err)
	}

	if deployment != nil {
		if confirm.DeploymentDeleted(deployment) {
			return nil
		}

		return errors2.MakeFailedError(fmt.Errorf("deployment not deleted"))
	}

	return nil
}

func UpdateDeployment(job *jobModels.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var update body.DeploymentUpdate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &update)
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	err = service.V1().Deployments().Update(id, &update)
	if err != nil {
		switch {
		case errors.Is(err, sErrors.DeploymentNotFoundErr):
			return errors2.MakeTerminatedError(err)
		case errors.Is(err, sErrors.NonUniqueFieldErr):
			return errors2.MakeTerminatedError(err)
		case errors.Is(err, sErrors.IngressHostInUseErr):
			return errors2.MakeTerminatedError(err)
		}

		return errors2.MakeFailedError(err)
	}

	err = deploymentModels.New().MarkUpdated(id)
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	return nil
}

func UpdateDeploymentOwner(job *jobModels.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var params body.DeploymentUpdateOwner
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	err = service.V1().Deployments().UpdateOwner(id, &params)
	if err != nil {
		if errors.Is(err, sErrors.DeploymentNotFoundErr) {
			return errors2.MakeTerminatedError(err)
		}

		return errors2.MakeFailedError(err)
	}

	return nil
}

func BuildDeployments(job *jobModels.Job) error {
	err := utils.AssertParameters(job, []string{"ids", "build"})
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	idsInt := job.Args["ids"].(primitive.A)
	ids := make([]string, len(idsInt))
	for idx, id := range idsInt {
		ids[idx] = id.(string)
	}

	var params body.DeploymentBuild
	err = mapstructure.Decode(job.Args["build"].(map[string]interface{}), &params)
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	var filtered []string
	for _, id := range ids {
		deleted, err := utils.DeploymentDeletedByID(id)
		if err != nil {
			return errors2.MakeTerminatedError(err)
		}

		if !deleted {
			filtered = append(filtered, id)
		}
	}

	if len(filtered) == 0 {
		return nil
	}

	err = service.V1().Deployments().Build(filtered, &params)
	if err != nil {
		return errors2.MakeFailedError(err)
	}

	return nil
}

func RepairDeployment(job *jobModels.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = service.V1().Deployments().Repair(id)
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	return nil
}

func CreateSM(job *jobModels.Job) error {
	err := utils.AssertParameters(job, []string{"id", "userId", "params"})
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	userID := job.Args["userId"].(string)

	var params smModels.CreateParams
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	err = service.V1().SMs().Create(id, userID, &params)
	if err != nil {
		// We always terminate these jobs, since rerunning it would cause a NonUniqueFieldErr
		return errors2.MakeTerminatedError(err)
	}

	return nil
}

func DeleteSM(job *jobModels.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = service.V1().SMs().Delete(id)
	if err != nil {
		return errors2.MakeFailedError(err)
	}

	return nil
}

func RepairSM(job *jobModels.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = service.V1().SMs().Repair(id)
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return errors2.MakeTerminatedError(err)
	}

	return nil
}

func RepairVM(job *jobModels.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return errors2.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = service.V1().VMs().Repair(id)
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
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

	err = service.V1().VMs().CreateSnapshot(vmID, &opts.CreateSnapshotOpts{System: &params})
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

	err = service.V1().VMs().CreateSnapshot(vmID, &opts.CreateSnapshotOpts{User: &params})
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

	err = service.V1().VMs().DeleteSnapshot(vmID, snapshotID)
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return errors2.MakeTerminatedError(err)
	}

	return nil
}
