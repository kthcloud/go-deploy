package job_execute

import (
	"context"
	"fmt"
	jobModel "go-deploy/models/sys/job"
	"go-deploy/pkg/jobs"
	"go-deploy/utils"
	"log"
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

			jobs.NewRunner(job).Run()

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

			jobs.NewRunner(job).Run()
			
		case <-ctx.Done():
			return
		}
	}
}
