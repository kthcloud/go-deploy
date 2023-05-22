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
		case jobModel.TypeCreateVM:
			go createVM(job)
		case jobModel.TypeDeleteVM:
			go deleteVM(job)
		case jobModel.TypeUpdateVM:
			go updateVM(job)
		case jobModel.TypeAttachGpuToVM:
			go attachGpuToVM(job)
		case jobModel.TypeDetachGpuFromVM:
			go detachGpuFromVM(job)
		case jobModel.TypeCreateDeployment:
			go createDeployment(job)
		case jobModel.TypeDeleteDeployment:
			go deleteDeployment(job)
		case jobModel.TypeUpdateDeployment:
			go updateDeployment(job)
		case jobModel.TypeBuildDeployment:
			go buildDeployment(job)
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
