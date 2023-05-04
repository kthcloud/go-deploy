package jobs

import (
	jobModel "go-deploy/models/sys/job"
	"go-deploy/pkg/app"
	"time"
)

func jobFetcher(ctx *app.Context) {
	for {
		if ctx.Stop {
			break
		}

		job, err := jobModel.GetNext()
		if err != nil {
			continue
		}

		if job == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		switch job.Type {
		case "createVm":
			go createVM(job)
		case "deleteVm":
			go deleteVM(job)
		case "updateVm":
			go updateVM(job)
		case "createDeployment":
			go createDeployment(job)
		case "deleteDeployment":
			go deleteDeployment(job)
		case "updateDeployment":
			go updateDeployment(job)
		}
	}
}

func failedJobFetcher(ctx *app.Context) {
	for {
		if ctx.Stop {
			break
		}

		time.Sleep(30 * time.Second)

		job, err := jobModel.GetNextFailed()
		if err != nil {
			continue
		}

		if job == nil {
			continue
		}

		switch job.Type {
		case "createVm":
			go createVM(job)
		case "deleteVm":
			go deleteVM(job)
		case "createDeployment":
			go createDeployment(job)
		case "deleteDeployment":
			go deleteDeployment(job)
		}
	}
}
