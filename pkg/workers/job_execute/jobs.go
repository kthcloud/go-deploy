package job_execute

import (
	"context"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"go-deploy/models/dto/body"
	deploymentModel "go-deploy/models/sys/deployment"
	storageManagerModel "go-deploy/models/sys/deployment/storage_manager"
	jobModel "go-deploy/models/sys/job"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/workers/confirm"
	"go-deploy/service/deployment_service"
	"go-deploy/service/vm_service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slices"
	"log"
	"strings"
	"time"
)

func makeTerminatedError(err error) error {
	return fmt.Errorf("terminated: %w", err)
}

func makeFailedError(err error) error {
	return fmt.Errorf("failed: %w", err)
}

func waitForJob(context context.Context, job *jobModel.Job, statuses []string) error {
	if len(statuses) == 0 {
		return nil
	}

	if slices.IndexFunc(statuses, func(s string) bool { return s == job.Status }) != -1 {
		return nil
	}

	for {
		select {
		case <-context.Done():
			return context.Err()
		default:
			var err error
			job, err = jobModel.New().GetByID(job.ID)
			if err != nil {
				return err
			}

			if slices.IndexFunc(statuses, func(s string) bool { return s == job.Status }) != -1 {
				return nil
			}

			time.Sleep(1 * time.Second)
		}
	}
}

func waitForJobs(context context.Context, jobs []jobModel.Job, statuses []string) error {
	for _, job := range jobs {
		err := waitForJob(context, &job, statuses)
		if err != nil {
			return err
		}
	}

	return nil
}

func deploymentDeleted(id string) (bool, error) {
	deleted, err := deploymentModel.New().IncludeDeletedResources().Deleted(id)
	if err != nil {
		return false, err
	}

	if deleted {
		return true, nil
	}

	beingDeleted, err := deploymentModel.New().DoingActivity(id, deploymentModel.ActivityBeingDeleted)
	if err != nil {
		return false, err
	}

	if beingDeleted {
		return true, nil
	}

	return false, nil
}

func vmDeleted(id string) (bool, error) {
	deleted, err := vmModel.New().IncludeDeletedResources().Deleted(id)
	if err != nil {
		return false, err
	}

	if deleted {
		return true, nil
	}

	beingDeleted, err := vmModel.New().DoingActivity(id, vmModel.ActivityBeingDeleted)
	if err != nil {
		return false, err
	}

	if beingDeleted {
		return true, nil
	}

	return false, nil
}

func createVM(job *jobModel.Job) error {
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

	deleted, err := vmDeleted(id)
	if err != nil {
		return makeTerminatedError(err)
	}

	if deleted {
		return makeTerminatedError(fmt.Errorf("vm is deleted"))
	}

	err = vm_service.Create(id, ownerID, &params)
	if err != nil {
		if strings.HasSuffix(err.Error(), "vm already exists for another user") {
			return makeTerminatedError(err)
		}

		return makeFailedError(err)
	}

	return nil
}

func deleteVM(job *jobModel.Job) error {
	err := assertParameters(job, []string{"id"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = vmModel.New().AddActivity(id, vmModel.ActivityBeingDeleted)
	if err != nil {
		return makeTerminatedError(err)
	}

	relatedJobs, err := jobModel.New().AddFilter(bson.D{{"id", bson.M{"$ne": job.ID}}}).GetByArgs(map[string]interface{}{"id": id})
	if err != nil {
		return makeTerminatedError(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	go func() {
		err = waitForJobs(ctx, relatedJobs, []string{jobModel.StatusCompleted, jobModel.StatusFailed, jobModel.StatusTerminated})
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

	err = vm_service.Delete(id)
	if err != nil {
		return makeFailedError(err)
	}

	// check if deleted, otherwise mark as failed and return to queue for retry
	vm, err := vmModel.New().GetByID(id)
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

func updateVM(job *jobModel.Job) error {
	err := assertParameters(job, []string{"id", "update"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var update body.VmUpdate
	err = mapstructure.Decode(job.Args["update"].(map[string]interface{}), &update)
	if err != nil {
		return makeTerminatedError(err)
	}

	deleted, err := vmDeleted(id)
	if err != nil {
		return makeTerminatedError(err)
	}

	if deleted {
		return makeTerminatedError(fmt.Errorf("vm is deleted"))
	}

	err = vm_service.Update(id, &update)
	if err != nil {
		if errors.Is(err, vm_service.NonUniqueFieldErr) {
			return makeTerminatedError(err)
		}

		return makeFailedError(err)
	}

	err = vmModel.New().MarkUpdated(id)
	if err != nil {
		return makeTerminatedError(err)
	}

	return nil
}

func attachGpuToVM(job *jobModel.Job) error {
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
	userID := job.Args["userId"].(string)
	leaseDuration := job.Args["leaseDuration"].(float64)

	deleted, err := vmDeleted(vmID)
	if err != nil {
		return makeTerminatedError(err)
	}

	if deleted {
		return makeTerminatedError(fmt.Errorf("vm is deleted"))
	}

	err = vm_service.AttachGPU(gpuIDs, vmID, userID, leaseDuration)
	if err != nil {
		return makeFailedError(err)
	}

	return nil
}

func detachGpuFromVM(job *jobModel.Job) error {
	err := assertParameters(job, []string{"id", "userId"})
	if err != nil {
		return makeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)
	userID := job.Args["userId"].(string)

	deleted, err := vmDeleted(vmID)
	if err != nil {
		return makeTerminatedError(err)
	}

	if deleted {
		return makeTerminatedError(fmt.Errorf("vm is deleted"))
	}

	err = vm_service.DetachGPU(vmID, userID)
	if err != nil {
		return makeFailedError(err)
	}

	return nil
}

func createDeployment(job *jobModel.Job) error {
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

	deleted, err := deploymentDeleted(id)
	if err != nil {
		return makeTerminatedError(err)
	}

	if deleted {
		return makeTerminatedError(fmt.Errorf("deployment is deleted"))
	}

	err = deployment_service.Create(id, ownerID, &params)
	if err != nil {
		if strings.HasSuffix(err.Error(), "deployment already exists for another user") {
			return makeTerminatedError(err)
		}

		return makeFailedError(err)
	}

	return nil
}

func deleteDeployment(job *jobModel.Job) error {
	err := assertParameters(job, []string{"id"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = deploymentModel.New().AddActivity(id, deploymentModel.ActivityBeingDeleted)
	if err != nil {
		return makeTerminatedError(err)
	}

	relatedJobs, err := jobModel.New().GetByArgs(bson.M{"args.id": id})
	if err != nil {
		return makeTerminatedError(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go func() {
		err = waitForJobs(ctx, relatedJobs, []string{jobModel.StatusCompleted, jobModel.StatusFailed, jobModel.StatusTerminated})
		if err != nil {
			log.Println("failed to wait for related jobs", id, ". details:", err)
		}
	}()

	select {
	case <-time.After(30 * time.Second):
		return makeTerminatedError(fmt.Errorf("timeout waiting for related jobs to finish"))
	default:
	}

	err = deployment_service.Delete(id)
	if err != nil {
		return makeFailedError(err)
	}

	// check if deleted, otherwise mark as failed and return to queue for retry
	deployment, err := deploymentModel.New().GetByID(id)
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

func updateDeployment(job *jobModel.Job) error {
	err := assertParameters(job, []string{"id", "update"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var update body.DeploymentUpdate
	err = mapstructure.Decode(job.Args["update"].(map[string]interface{}), &update)
	if err != nil {
		return makeTerminatedError(err)
	}

	deleted, err := deploymentDeleted(id)
	if err != nil {
		return makeTerminatedError(err)
	}

	if deleted {
		return makeTerminatedError(fmt.Errorf("deployment is deleted"))
	}

	err = deployment_service.Update(id, &update)
	if err != nil {
		if errors.Is(err, deployment_service.NonUniqueFieldErr) {
			return makeTerminatedError(err)
		}

		return makeFailedError(err)
	}

	err = deploymentModel.New().MarkUpdated(id)
	if err != nil {
		return makeTerminatedError(err)
	}

	return nil
}

func buildDeployments(job *jobModel.Job) error {
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
		deleted, err := deploymentDeleted(id)
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

	err = deployment_service.Build(filtered, &params)
	if err != nil {
		return makeFailedError(err)
	}

	return nil
}

func repairDeployment(job *jobModel.Job) error {
	err := assertParameters(job, []string{"id"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	deleted, err := deploymentDeleted(id)
	if err != nil {
		return makeTerminatedError(err)
	}

	if deleted {
		return makeTerminatedError(fmt.Errorf("deployment is deleted"))
	}

	err = deployment_service.Repair(id)
	if err != nil {
		return makeTerminatedError(err)
	}

	return nil
}

func createStorageManager(job *jobModel.Job) error {
	err := assertParameters(job, []string{"id", "params"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var params storageManagerModel.CreateParams
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return makeTerminatedError(err)
	}

	err = deployment_service.CreateStorageManager(id, &params)
	if err != nil {
		if errors.Is(err, deployment_service.StorageManagerAlreadyExistsErr) {
			return makeTerminatedError(err)
		}

		return makeFailedError(err)
	}

	return nil
}

func deleteStorageManager(job *jobModel.Job) error {
	err := assertParameters(job, []string{"id"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = deployment_service.DeleteStorageManager(id)
	if err != nil {
		return makeFailedError(err)
	}

	return nil
}

func repairStorageManager(job *jobModel.Job) error {
	err := assertParameters(job, []string{"id"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = deployment_service.RepairStorageManager(id)
	if err != nil {
		return makeTerminatedError(err)
	}

	return nil
}

func repairVM(job *jobModel.Job) error {
	err := assertParameters(job, []string{"id"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	deleted, err := vmDeleted(id)
	if err != nil {
		return makeTerminatedError(err)
	}

	if deleted {
		return makeTerminatedError(fmt.Errorf("vm is deleted"))
	}

	err = vm_service.Repair(id)
	if err != nil {
		return makeTerminatedError(err)
	}

	return nil
}

func repairGPUs(job *jobModel.Job) error {
	err := assertParameters(job, []string{})
	if err != nil {
		return makeTerminatedError(err)
	}

	err = vm_service.RepairGPUs()
	if err != nil {
		return makeTerminatedError(err)
	}

	return nil
}

func createSnapshot(job *jobModel.Job) error {
	err := assertParameters(job, []string{"id", "name", "userCreated"})
	if err != nil {
		return makeTerminatedError(err)
	}

	vmID := job.Args["id"].(string)
	name := job.Args["name"].(string)
	userCreated := job.Args["userCreated"].(bool)

	deleted, err := vmDeleted(vmID)
	if err != nil {
		return makeTerminatedError(err)
	}

	if deleted {
		return makeTerminatedError(fmt.Errorf("vm is deleted"))
	}

	err = vm_service.CreateSnapshot(vmID, name, userCreated)
	if err != nil {
		return makeTerminatedError(err)
	}

	return nil
}

func applySnapshot(job *jobModel.Job) error {
	err := assertParameters(job, []string{"id", "snapshotId"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	snapshotID := job.Args["snapshotId"].(string)

	deleted, err := vmDeleted(id)
	if err != nil {
		return makeTerminatedError(err)
	}

	if deleted {
		return makeTerminatedError(fmt.Errorf("vm is deleted"))
	}

	err = vm_service.ApplySnapshot(id, snapshotID)
	if err != nil {
		return makeTerminatedError(err)
	}

	return nil
}
