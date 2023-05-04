package jobs

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"go-deploy/models/dto/body"
	jobModel "go-deploy/models/sys/job"
	"go-deploy/pkg/app"
	"go-deploy/service/deployment_service"
	"go-deploy/service/vm_service"
	"log"
)

func assertParameters(job *jobModel.Job, params []string) error {
	for _, param := range params {
		if _, ok := job.Args[param]; !ok {
			return fmt.Errorf("missing parameter: %s", param)
		}
	}

	return nil
}

func createVM(job *jobModel.Job) {
	err := assertParameters(job, []string{"id", "ownerId", "params"})
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, []string{err.Error()})
		return
	}

	id := job.Args["id"].(string)
	ownerID := job.Args["ownerId"].(string)
	var params body.VmCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, []string{err.Error()})
		return
	}

	err = vm_service.Create(id, params.Name, params.SshPublicKey, ownerID)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, []string{err.Error()})
		return
	}

	_ = jobModel.MarkCompleted(job.ID)
}

func deleteVM(job *jobModel.Job) {
	err := assertParameters(job, []string{"name"})
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, []string{err.Error()})
		return
	}

	name := job.Args["name"].(string)

	err = vm_service.Delete(name)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, []string{err.Error()})
		return
	}

	_ = jobModel.MarkCompleted(job.ID)
}

func updateVM(job *jobModel.Job) {
	err := assertParameters(job, []string{"name", "update"})
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, []string{err.Error()})
		return
	}

	name := job.Args["name"].(string)
	var update body.VmUpdate
	err = mapstructure.Decode(job.Args["update"].(map[string]interface{}), &update)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, []string{err.Error()})
		return
	}

	err = vm_service.Update(name, &update)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, []string{err.Error()})
		return
	}

	_ = jobModel.MarkCompleted(job.ID)
}

func createDeployment(job *jobModel.Job) {
	err := assertParameters(job, []string{"id", "ownerId", "params"})
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, []string{err.Error()})
		return
	}

	id := job.Args["id"].(string)
	ownerID := job.Args["ownerId"].(string)
	var params body.DeploymentCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, []string{err.Error()})
		return
	}

	err = deployment_service.Create(id, ownerID, &params)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, []string{err.Error()})
		return
	}

	_ = jobModel.MarkCompleted(job.ID)
}

func deleteDeployment(job *jobModel.Job) {
	err := assertParameters(job, []string{"name"})
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, []string{err.Error()})
		return
	}

	name := job.Args["name"].(string)

	err = deployment_service.Delete(name)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, []string{err.Error()})
		return
	}

	_ = jobModel.MarkCompleted(job.ID)
}

func updateDeployment(job *jobModel.Job) {
	err := assertParameters(job, []string{"name", "update"})
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, []string{err.Error()})
		return
	}

	name := job.Args["name"].(string)
	var update body.DeploymentUpdate
	err = mapstructure.Decode(job.Args["update"].(map[string]interface{}), &update)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, []string{err.Error()})
		return
	}
	err = deployment_service.Update(name, &update)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, []string{err.Error()})
		return
	}

	_ = jobModel.MarkCompleted(job.ID)
}

func Setup(ctx *app.Context) {
	log.Println("starting job workers")
	go jobFetcher(ctx)
	go failedJobFetcher(ctx)
}
