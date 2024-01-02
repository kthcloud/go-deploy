package jobs

import (
	"context"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"go-deploy/models/dto/body"
	deploymentModels "go-deploy/models/sys/deployment"
	jobModels "go-deploy/models/sys/job"
	smModels "go-deploy/models/sys/sm"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/pkg/workers/confirm"
	"go-deploy/service/deployment_service"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/sm_service"
	"go-deploy/service/vm_service"
	"go-deploy/service/vm_service/client"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"time"
)

func CreateVM(job *jobModels.Job) error {
	err := assertParameters(job, []string{"id", "ownerId", "params"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	ownerID := job.Args["ownerId"].(string)
	var params body.VmCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return makeTerminatedError(err)
	}

	err = vm_service.New().Create(id, ownerID, &params)
	if err != nil {
		// If there was some error, we trigger a repair, since rerunning it would cause a NonUniqueFieldErr
		_ = vm_service.New().Repair(id)
		return makeTerminatedError(err)
	}

	return nil
}

func DeleteVM(job *jobModels.Job) error {
	err := assertParameters(job, []string{"id"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = vmModels.New().AddActivity(id, vmModels.ActivityBeingDeleted)
	if err != nil {
		return makeTerminatedError(err)
	}

	relatedJobs, err := jobModels.New().
		ExcludeScheduled().
		ExcludeTypes(jobModels.TypeDeleteVM).
		ExcludeStatus(jobModels.StatusTerminated, jobModels.StatusCompleted).
		ExcludeIDs(job.ID).
		FilterArgs("id", id).
		List()
	if err != nil {
		return makeTerminatedError(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	go func() {
		err = waitForJobs(ctx, relatedJobs, []string{jobModels.StatusCompleted, jobModels.StatusTerminated})
		if err != nil {
			log.Println("failed to wait for related jobs", id, ". details:", err)
		}
		cancel()
	}()

	select {
	case <-time.After(300 * time.Second):
		return makeTerminatedError(fmt.Errorf("timeout waiting for related jobs to finish"))
	case <-ctx.Done():
	}

	err = vm_service.New().Delete(id)
	if err != nil {
		if !errors.Is(err, sErrors.VmNotFoundErr) {
			return makeFailedError(err)
		}
	}

	// Check if deleted, otherwise mark as failed and return to queue for retry
	vm, err := vmModels.New().GetByID(id)
	if err != nil {
		return makeFailedError(err)
	}

	if vm != nil {
		if confirm.VmDeleted(vm) {
			return nil
		}

		return makeFailedError(fmt.Errorf("vm not deleted"))
	}

	return nil
}

func UpdateVM(job *jobModels.Job) error {
	err := assertParameters(job, []string{"id", "params"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var update body.VmUpdate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &update)
	if err != nil {
		return makeTerminatedError(err)
	}

	err = vm_service.New().Update(id, &update)
	if err != nil {
		switch {
		case errors.Is(err, sErrors.VmNotFoundErr):
			return makeTerminatedError(err)
		case errors.Is(err, sErrors.NonUniqueFieldErr):
			return makeTerminatedError(err)
		case errors.Is(err, sErrors.IngressHostInUseErr):
			return makeTerminatedError(err)
		case errors.Is(err, sErrors.PortInUseErr):
			return makeTerminatedError(err)
		case errors.Is(err, sErrors.NoPortsAvailableErr):
			return makeTerminatedError(err)
		case errors.Is(err, sErrors.SnapshotNotFoundErr):
			return makeTerminatedError(err)
		}

		return makeFailedError(err)
	}

	err = vmModels.New().MarkUpdated(id)
	if err != nil {
		return makeTerminatedError(err)
	}

	return nil
}

func UpdateVmOwner(job *jobModels.Job) error {
	err := assertParameters(job, []string{"id", "params"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var params body.VmUpdateOwner
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return makeTerminatedError(err)
	}

	err = vm_service.New().UpdateOwner(id, &params)
	if err != nil {
		return makeFailedError(err)
	}

	return nil
}

func AttachGpuToVM(job *jobModels.Job) error {
	err := assertParameters(job, []string{"id", "gpuIds", "userId", "leaseDuration"})
	if err != nil {
		return makeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)
	var gpuIDs []string
	err = mapstructure.Decode(job.Args["gpuIds"].(interface{}), &gpuIDs)
	if err != nil {
		return makeTerminatedError(err)
	}
	leaseDuration := job.Args["leaseDuration"].(float64)

	// We keep this field to know who requested the gpu attachment
	_ = job.Args["userId"].(string)

	err = vm_service.New().AttachGPU(vmID, gpuIDs, leaseDuration)
	if err != nil {
		if errors.Is(err, sErrors.GpuNotFoundErr) {
			return makeTerminatedError(err)
		}

		if errors.Is(err, sErrors.VmNotFoundErr) {
			return makeTerminatedError(err)
		}

		return makeFailedError(err)
	}

	return nil
}

func DetachGpuFromVM(job *jobModels.Job) error {
	err := assertParameters(job, []string{"id"})
	if err != nil {
		return makeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)

	err = vm_service.New().DetachGPU(vmID)
	if err != nil {
		return makeFailedError(err)
	}

	return nil
}

func CreateDeployment(job *jobModels.Job) error {
	err := assertParameters(job, []string{"id", "ownerId", "params"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	ownerID := job.Args["ownerId"].(string)
	var params body.DeploymentCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return makeTerminatedError(err)
	}

	err = deployment_service.New().Create(id, ownerID, &params)
	if err != nil {
		// If there was some error, we trigger a repair, since rerunning it would cause a NonUniqueFieldErr
		_ = deployment_service.New().Repair(id)
		return makeTerminatedError(err)
	}

	return nil
}

func DeleteDeployment(job *jobModels.Job) error {
	err := assertParameters(job, []string{"id"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = deploymentModels.New().AddActivity(id, deploymentModels.ActivityBeingDeleted)
	if err != nil {
		return makeTerminatedError(err)
	}

	relatedJobs, err := jobModels.New().FilterArgs("id", id).List()
	if err != nil {
		return makeTerminatedError(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go func() {
		err = waitForJobs(ctx, relatedJobs, []string{jobModels.StatusCompleted, jobModels.StatusTerminated})
		if err != nil {
			log.Println("failed to wait for related jobs", id, ". details:", err)
		}
	}()

	select {
	case <-time.After(30 * time.Second):
		return makeTerminatedError(fmt.Errorf("timeout waiting for related jobs to finish"))
	default:
	}

	err = deployment_service.New().Delete(id)
	if err != nil {
		if !errors.Is(err, sErrors.DeploymentNotFoundErr) {
			return makeFailedError(err)
		}
	}

	// Check if deleted, otherwise mark as failed and return to queue for retry
	deployment, err := deploymentModels.New().GetByID(id)
	if err != nil {
		return makeFailedError(err)
	}

	if deployment != nil {
		if confirm.DeploymentDeleted(deployment) {
			return nil
		}

		return makeFailedError(fmt.Errorf("deployment not deleted"))
	}

	return nil
}

func UpdateDeployment(job *jobModels.Job) error {
	err := assertParameters(job, []string{"id", "params"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var update body.DeploymentUpdate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &update)
	if err != nil {
		return makeTerminatedError(err)
	}

	err = deployment_service.New().Update(id, &update)
	if err != nil {
		switch {
		case errors.Is(err, sErrors.DeploymentNotFoundErr):
			return makeTerminatedError(err)
		case errors.Is(err, sErrors.NonUniqueFieldErr):
			return makeTerminatedError(err)
		case errors.Is(err, sErrors.IngressHostInUseErr):
			return makeTerminatedError(err)
		}

		return makeFailedError(err)
	}

	err = deploymentModels.New().MarkUpdated(id)
	if err != nil {
		return makeTerminatedError(err)
	}

	return nil
}

func UpdateDeploymentOwner(job *jobModels.Job) error {
	err := assertParameters(job, []string{"id", "params"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var params body.DeploymentUpdateOwner
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return makeTerminatedError(err)
	}

	err = deployment_service.New().UpdateOwner(id, &params)
	if err != nil {
		if errors.Is(err, sErrors.DeploymentNotFoundErr) {
			return makeTerminatedError(err)
		}

		return makeFailedError(err)
	}

	return nil
}

func BuildDeployments(job *jobModels.Job) error {
	err := assertParameters(job, []string{"ids", "build"})
	if err != nil {
		return makeTerminatedError(err)
	}

	idsInt := job.Args["ids"].(primitive.A)
	ids := make([]string, len(idsInt))
	for idx, id := range idsInt {
		ids[idx] = id.(string)
	}

	var params body.DeploymentBuild
	err = mapstructure.Decode(job.Args["build"].(map[string]interface{}), &params)
	if err != nil {
		return makeTerminatedError(err)
	}

	var filtered []string
	for _, id := range ids {
		deleted, err := deploymentDeletedByID(id)
		if err != nil {
			return makeTerminatedError(err)
		}

		if !deleted {
			filtered = append(filtered, id)
		}
	}

	if len(filtered) == 0 {
		return nil
	}

	err = deployment_service.New().Build(filtered, &params)
	if err != nil {
		return makeFailedError(err)
	}

	return nil
}

func RepairDeployment(job *jobModels.Job) error {
	err := assertParameters(job, []string{"id"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = deployment_service.New().Repair(id)
	if err != nil {
		return makeTerminatedError(err)
	}

	return nil
}

func CreateSM(job *jobModels.Job) error {
	err := assertParameters(job, []string{"id", "userId", "params"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	userID := job.Args["userId"].(string)

	var params smModels.CreateParams
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return makeTerminatedError(err)
	}

	err = sm_service.New().Create(id, userID, &params)
	if err != nil {
		// We always terminate these jobs, since rerunning it would cause a NonUniqueFieldErr
		return makeTerminatedError(err)
	}

	return nil
}

func DeleteSM(job *jobModels.Job) error {
	err := assertParameters(job, []string{"id"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = sm_service.New().Delete(id)
	if err != nil {
		return makeFailedError(err)
	}

	return nil
}

func RepairSM(job *jobModels.Job) error {
	err := assertParameters(job, []string{"id"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = sm_service.New().Repair(id)
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return makeTerminatedError(err)
	}

	return nil
}

func RepairVM(job *jobModels.Job) error {
	err := assertParameters(job, []string{"id"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = vm_service.New().Repair(id)
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return makeTerminatedError(err)
	}

	return nil
}

func CreateSystemSnapshot(job *jobModels.Job) error {
	err := assertParameters(job, []string{"id", "params"})
	if err != nil {
		return makeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)
	var params vmModels.CreateSnapshotParams
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return makeTerminatedError(err)
	}

	err = vm_service.New().CreateSnapshot(vmID, &client.CreateSnapshotOptions{System: &params})
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return makeTerminatedError(err)
	}

	return nil
}

func CreateUserSnapshot(job *jobModels.Job) error {
	err := assertParameters(job, []string{"id", "params"})
	if err != nil {
		return makeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)
	var params body.VmSnapshotCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return makeTerminatedError(err)
	}

	err = vm_service.New().CreateSnapshot(vmID, &client.CreateSnapshotOptions{User: &params})
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return makeTerminatedError(err)
	}

	return nil
}

func DeleteSnapshot(job *jobModels.Job) error {
	err := assertParameters(job, []string{"id", "snapshotId"})
	if err != nil {
		return makeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)
	snapshotID := job.Args["snapshotId"].(string)

	err = vm_service.New().DeleteSnapshot(vmID, snapshotID)
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return makeTerminatedError(err)
	}

	return nil
}
