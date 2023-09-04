package job_execute

import (
	"context"
	"errors"
	"fmt"
	jobModel "go-deploy/models/sys/job"
	"go-deploy/utils"
	"log"
	"strings"
	"time"
)

func jobFetcher(ctx context.Context) {
	defer log.Println("jobFetcher stopped")

	for {
		select {
		case <-time.After(100 * time.Millisecond):
			job, err := jobModel.New().GetNext()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error fetching next job. details: %w", err))
				continue
			}

			if job == nil {
				continue
			}

			err = startJob(job)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error starting failed job. details: %w", err))
				err = jobModel.New().MarkTerminated(job.ID, err.Error())
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error marking failed job as terminated. details: %w", err))
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
			job, err := jobModel.New().GetNextFailed()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error fetching next failed job. details: %w", err))
				continue
			}

			if job == nil {
				continue
			}

			err = startJob(job)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error starting failed job. details: %w", err))
				err = jobModel.New().MarkTerminated(job.ID, err.Error())
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error marking failed job as terminated. details: %w", err))
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
		go wrapper(createVM, job)
	case jobModel.TypeDeleteVM:
		go wrapper(deleteVM, job)
	case jobModel.TypeUpdateVM:
		go wrapper(updateVM, job)
	case jobModel.TypeAttachGpuToVM:
		go wrapper(attachGpuToVM, job)
	case jobModel.TypeDetachGpuFromVM:
		go wrapper(detachGpuFromVM, job)
	case jobModel.TypeCreateDeployment:
		go wrapper(createDeployment, job)
	case jobModel.TypeDeleteDeployment:
		go wrapper(deleteDeployment, job)
	case jobModel.TypeUpdateDeployment:
		go wrapper(updateDeployment, job)
	case jobModel.TypeBuildDeployment:
		go wrapper(buildDeployment, job)
	case jobModel.TypeRepairDeployment:
		go wrapper(repairDeployment, job)
	case jobModel.TypeCreateStorageManager:
		go wrapper(createStorageManager, job)
	case jobModel.TypeDeleteStorageManager:
		go wrapper(deleteStorageManager, job)
	case jobModel.TypeRepairStorageManager:
		go wrapper(repairStorageManager, job)
	case jobModel.TypeRepairVM:
		go wrapper(repairVM, job)
	case jobModel.TypeRepairGPUs:
		go wrapper(repairGPUs, job)
	case jobModel.TypeCreateSnapshot:
		go wrapper(createSnapshot, job)
	case jobModel.TypeApplySnapshot:
		go wrapper(applySnapshot, job)
	default:
		return fmt.Errorf("unknown job type: %s", job.Type)
	}
	return nil
}

func wrapper(fn func(job *jobModel.Job) error, job *jobModel.Job) {
	err := fn(job)
	if err != nil {
		if strings.HasPrefix(err.Error(), "failed") {
			err = errors.Unwrap(err)
			utils.PrettyPrintError(fmt.Errorf("failed job (%s). details: %w", job.Type, err))

			err := jobModel.New().MarkFailed(job.ID, err.Error())
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error marking job as failed. details: %w", err))
				return
			}
		} else if strings.HasPrefix(err.Error(), "terminated") {
			err = errors.Unwrap(err)
			utils.PrettyPrintError(fmt.Errorf("terminated job (%s). details: %w", job.Type, err))

			err := jobModel.New().MarkTerminated(job.ID, err.Error())
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error marking job as terminated. details: %w", err))
				return
			}
		} else {
			utils.PrettyPrintError(fmt.Errorf("error executing job (%s). details: %w", job.Type, err))

			err := jobModel.New().MarkFailed(job.ID, err.Error())
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error marking job as failed. details: %w", err))
				return
			}
		}
	} else {
		err := jobModel.New().MarkCompleted(job.ID)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("error marking job as completed. details: %w", err))
			return
		}
	}
}
