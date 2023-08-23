package job_execute

import (
	"context"
	"fmt"
	jobModel "go-deploy/models/sys/job"
	"log"
	"time"
)

func jobFetcher(ctx context.Context) {
	defer log.Println("jobFetcher stopped")

	for {
		select {
		case <-time.After(100 * time.Millisecond):
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

		case <-ctx.Done():
			return
		}
	}
}

func failedJobFetcher(ctx context.Context) {
	defer log.Println("failedJobFetcher stopped")

	for {
		select {
		case <-time.After(30 * time.Second):
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
		case <-ctx.Done():
			return
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
	case jobModel.TypeRepairDeployment:
		go repairDeployment(job)
	case jobModel.TypeCreateStorageManager:
		go createStorageManager(job)
	case jobModel.TypeDeleteStorageManager:
		go deleteStorageManager(job)
	case jobModel.TypeRepairStorageManager:
		go repairStorageManager(job)
	case jobModel.TypeRepairVM:
		go repairVM(job)
	case jobModel.TypeRepairGPUs:
		go repairGPUs(job)
	case jobModel.TypeCreateSnapshot:
		go createSnapshot(job)
	case jobModel.TypeApplySnapshot:
		go applySnapshot(job)
	default:
		return fmt.Errorf("unknown job type: %s", job.Type)
	}
	return nil
}
