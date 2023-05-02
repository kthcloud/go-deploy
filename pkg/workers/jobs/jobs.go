package jobs

import (
	"fmt"
	"go-deploy/models/dto/body"
	jobModel "go-deploy/models/job"
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
	err := assertParameters(job, []string{"id", "name", "sshPublicKey", "ownerId"})
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, []string{err.Error()})
		return
	}

	id := job.Args["id"].(string)
	name := job.Args["name"].(string)
	sshPublicKey := job.Args["sshPublicKey"].(string)
	ownerID := job.Args["ownerId"].(string)

	err = vm_service.Create(id, name, sshPublicKey, ownerID)
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
	update := job.Args["update"].(body.VmUpdate)

	err = vm_service.Update(name, &update)
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, []string{err.Error()})
		return
	}

	_ = jobModel.MarkCompleted(job.ID)
}

func createDeployment(job *jobModel.Job) {
	err := assertParameters(job, []string{"id", "name", "ownerId"})
	if err != nil {
		_ = jobModel.MarkFailed(job.ID, []string{err.Error()})
		return
	}

	id := job.Args["id"].(string)
	name := job.Args["name"].(string)
	ownerID := job.Args["ownerId"].(string)

	err = deployment_service.Create(id, name, ownerID)
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

	//name := job.Args["name"].(string)
	//update := job.Args["update"].(body.DeploymentUpdate)
	//
	//err = deployment_service.Update(name, &update)
	//if err != nil {
	//	_ = jobModel.MarkFailed(job.ID, []string{err.Error()})
	//	return
	//}

	_ = jobModel.MarkCompleted(job.ID)
}

func Setup(ctx *app.Context) {
	log.Println("starting job workers")
	go jobFetcher(ctx)
	go failedJobFetcher(ctx)
}
