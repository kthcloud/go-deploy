package job_execute

import (
	"github.com/mitchellh/mapstructure"
	"go-deploy/models/dto/body"
	jobModel "go-deploy/models/sys/job"
	"go-deploy/service/deployment_service"
	"go-deploy/service/vm_service"
)

func createVM(job *jobModel.Job) {
	err := assertParameters(job, []string{"id", "ownerId", "params"})
	if err != nil {
		_ = jobModel.MarkTerminated(job.ID, err.Error())
		return
	}

	id := job.Args["id"].(string)
	ownerID := job.Args["ownerId"].(string)
	var params body.VmCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		_ = jobModel.MarkTerminated(job.ID, err.Error())
		return
	}

	err = vm_service.Create(id, ownerID, &params)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, err.Error())
		return
	}

	_ = jobModel.MarkCompleted(job.ID)
}

func deleteVM(job *jobModel.Job) {
	err := assertParameters(job, []string{"name"})
	if err != nil {
		_ = jobModel.MarkTerminated(job.ID, err.Error())
		return
	}

	name := job.Args["name"].(string)

	err = vm_service.Delete(name)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, err.Error())
		return
	}

	_ = jobModel.MarkCompleted(job.ID)
}

func updateVM(job *jobModel.Job) {
	err := assertParameters(job, []string{"id", "update"})
	if err != nil {
		_ = jobModel.MarkTerminated(job.ID, err.Error())
		return
	}

	id := job.Args["id"].(string)
	var update body.VmUpdate
	err = mapstructure.Decode(job.Args["update"].(map[string]interface{}), &update)
	if err != nil {
		_ = jobModel.MarkTerminated(job.ID, err.Error())
		return
	}

	err = vm_service.Update(id, &update)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, err.Error())
		return
	}

	_ = jobModel.MarkCompleted(job.ID)
}

func attachGpuToVM(job *jobModel.Job) {
	err := assertParameters(job, []string{"id", "gpuIds", "userId"})
	if err != nil {
		_ = jobModel.MarkTerminated(job.ID, err.Error())
		return
	}

	vmID := job.Args["id"].(string)
	var gpuIDs []string
	err = mapstructure.Decode(job.Args["gpuIds"].(interface{}), &gpuIDs)
	if err != nil {
		_ = jobModel.MarkTerminated(job.ID, err.Error())
		return
	}
	userID := job.Args["userId"].(string)

	err = vm_service.AttachGPU(gpuIDs, vmID, userID)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, err.Error())
		return
	}

	_ = jobModel.MarkCompleted(job.ID)
}

func detachGpuFromVM(job *jobModel.Job) {
	err := assertParameters(job, []string{"id", "userId"})
	if err != nil {
		_ = jobModel.MarkTerminated(job.ID, err.Error())
		return
	}

	vmID := job.Args["id"].(string)
	userID := job.Args["userId"].(string)

	err = vm_service.DetachGPU(vmID, userID)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, err.Error())
		return
	}

	_ = jobModel.MarkCompleted(job.ID)
}

func createDeployment(job *jobModel.Job) {
	err := assertParameters(job, []string{"id", "ownerId", "params"})
	if err != nil {
		_ = jobModel.MarkTerminated(job.ID, err.Error())
		return
	}

	id := job.Args["id"].(string)
	ownerID := job.Args["ownerId"].(string)
	var params body.DeploymentCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		_ = jobModel.MarkTerminated(job.ID, err.Error())
		return
	}

	err = deployment_service.Create(id, ownerID, &params)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, err.Error())
		return
	}

	_ = jobModel.MarkCompleted(job.ID)
}

func deleteDeployment(job *jobModel.Job) {
	err := assertParameters(job, []string{"name"})
	if err != nil {
		_ = jobModel.MarkTerminated(job.ID, err.Error())
		return
	}

	name := job.Args["name"].(string)

	err = deployment_service.Delete(name)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, err.Error())
		return
	}

	_ = jobModel.MarkCompleted(job.ID)
}

func updateDeployment(job *jobModel.Job) {
	err := assertParameters(job, []string{"name", "update"})
	if err != nil {
		_ = jobModel.MarkTerminated(job.ID, err.Error())
		return
	}

	name := job.Args["name"].(string)
	var update body.DeploymentUpdate
	err = mapstructure.Decode(job.Args["update"].(map[string]interface{}), &update)
	if err != nil {
		_ = jobModel.MarkTerminated(job.ID, err.Error())
		return
	}

	err = deployment_service.Update(name, &update)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, err.Error())
		return
	}

	_ = jobModel.MarkCompleted(job.ID)
}

func buildDeployment(job *jobModel.Job) {
	err := assertParameters(job, []string{"id", "build"})
	if err != nil {
		_ = jobModel.MarkTerminated(job.ID, err.Error())
		return
	}

	id := job.Args["id"].(string)
	var params body.DeploymentBuild
	err = mapstructure.Decode(job.Args["build"].(map[string]interface{}), &params)
	if err != nil {
		_ = jobModel.MarkTerminated(job.ID, err.Error())
		return
	}

	err = deployment_service.Build(id, &params)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, err.Error())
		return
	}

	_ = jobModel.MarkCompleted(job.ID)
}

func repairDeployment(job *jobModel.Job) {
	err := assertParameters(job, []string{"id"})
	if err != nil {
		_ = jobModel.MarkTerminated(job.ID, err.Error())
		return
	}

	id := job.Args["id"].(string)

	err = deployment_service.Repair(id)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, err.Error())
		return
	}

	_ = jobModel.MarkCompleted(job.ID)
}
