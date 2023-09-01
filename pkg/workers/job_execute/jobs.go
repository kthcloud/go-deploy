package job_execute

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"go-deploy/models/dto/body"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/deployment/storage_manager"
	jobModel "go-deploy/models/sys/job"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/service/deployment_service"
	"go-deploy/service/vm_service"
	"strings"
)

func makeTerminatedError(err error) error {
	return fmt.Errorf("terminated: %w", err)
}

func makeFailedError(err error) error {
	return fmt.Errorf("failed: %w", err)
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
	err := assertParameters(job, []string{"name"})
	if err != nil {
		return makeTerminatedError(err)
	}

	name := job.Args["name"].(string)

	err = vm_service.Delete(name)
	if err != nil {
		return makeFailedError(err)
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

	err = vm_service.Update(id, &update)
	if err != nil {
		return makeFailedError(err)
	}

	err = vmModel.MarkUpdated(id)
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
	err := assertParameters(job, []string{"name"})
	if err != nil {
		return makeTerminatedError(err)
	}

	name := job.Args["name"].(string)

	err = deployment_service.Delete(name)
	if err != nil {
		return makeFailedError(err)
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

	err = deployment_service.Update(id, &update)
	if err != nil {
		return makeFailedError(err)
	}

	err = deploymentModel.MarkUpdated(id)
	if err != nil {
		return makeTerminatedError(err)
	}

	return nil
}

func buildDeployment(job *jobModel.Job) error {
	err := assertParameters(job, []string{"id", "build"})
	if err != nil {
		return makeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var params body.DeploymentBuild
	err = mapstructure.Decode(job.Args["build"].(map[string]interface{}), &params)
	if err != nil {
		return makeTerminatedError(err)
	}

	err = deployment_service.Build(id, &params)
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
	var params storage_manager.CreateParams
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return makeTerminatedError(err)
	}

	err = deployment_service.CreateStorageManager(id, &params)
	if err != nil {
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

	id := job.Args["id"].(string)
	name := job.Args["name"].(string)
	userCreated := job.Args["userCreated"].(bool)

	err = vm_service.CreateSnapshot(id, name, userCreated)
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

	err = vm_service.ApplySnapshot(id, snapshotID)
	if err != nil {
		return makeTerminatedError(err)
	}

	return nil
}
