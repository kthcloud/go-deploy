package jobs

import (
	"fmt"
	jobModel "go-deploy/models/sys/job"
	"go-deploy/pkg/app"
	"log"
	"time"
)

func jobFetcher(ctx *app.Context) {
	for {
		if ctx.Stop {
			break
		}

		time.Sleep(100 * time.Millisecond)

		job, err := jobModel.GetNext()
		if err != nil {
			log.Println("error fetching next job. details: ", err)
			continue
		}

		if job == nil {
			continue
		}

		err = startJob(job)
		if err != nil {
			log.Println("error starting failed job. details: ", err)
			err = jobModel.MarkTerminated(job.ID, err.Error())
			if err != nil {
				log.Println("error marking failed job as terminated. details: ", err)
				return
			}
			continue
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
			log.Println("error fetching next failed job. details: ", err)
			continue
		}

		if job == nil {
			continue
		}

		err = startJob(job)
		if err != nil {
			log.Println("error starting failed job. details: ", err)
			err = jobModel.MarkTerminated(job.ID, err.Error())
			if err != nil {
				log.Println("error marking failed job as terminated. details: ", err)
				return
			}
			continue
		}
	}
}

func startJob(job *jobModel.Job) error {
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
	default:
		return fmt.Errorf("unknown job type: %s", job.Type)
	}
	return nil
}
