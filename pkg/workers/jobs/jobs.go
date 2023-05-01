package jobs

import (
	"fmt"
	jobModel "go-deploy/models/job"
	"go-deploy/pkg/app"
	"go-deploy/service/deployment_service"
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

func createDeployment(job *jobModel.Job) {
	err := assertParameters(job, []string{"id", "name", "ownerId"})
	if err != nil {
		_ = jobModel.FailJob(job.ID, []string{err.Error()})
		return
	}

	id := job.Args["id"].(string)
	name := job.Args["name"].(string)
	ownerID := job.Args["ownerId"].(string)

	err = deployment_service.Create(id, name, ownerID)
	if err != nil {
		_ = jobModel.FailJob(job.ID, []string{err.Error()})
		return
	}

	_ = jobModel.CompleteJob(job.ID)
}

func Setup(ctx *app.Context) {
	log.Println("starting job workers")
	go jobFetcher(ctx)
}
