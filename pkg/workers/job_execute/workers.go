package job_execute

import (
	"context"
	"fmt"
	jobModels "go-deploy/models/sys/job"
	"go-deploy/pkg/jobs"
	"go-deploy/pkg/workers"
	"go-deploy/utils"
	"time"
)

func jobFetcher(ctx context.Context) {
	defer workers.OnStop("jobFetcher")

	for {
		select {
		case <-time.After(1 * time.Second):
			workers.ReportUp("jobFetcher")

		case <-time.After(100 * time.Millisecond):
			job, err := jobModels.New().GetNext()
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
	defer workers.OnStop("failedJobFetcher")

	for {
		select {
		case <-time.After(1 * time.Second):
			workers.ReportUp("failedJobFetcher")

		case <-time.After(30 * time.Second):
			job, err := jobModels.New().GetNextFailed()
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
