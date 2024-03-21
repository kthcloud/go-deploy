package job_execute

import (
	"context"
	"fmt"
	"go-deploy/pkg/db/resources/job_repo"
	"go-deploy/pkg/jobs"
	"go-deploy/pkg/workers"
	"go-deploy/utils"
	"time"
)

// jobFetcher is a worker that fetches jobs from the database and runs them.
func jobFetcher(ctx context.Context) {
	defer workers.OnStop("jobFetcher")

	reportTick := time.Tick(1 * time.Second)
	tick := time.Tick(100 * time.Millisecond)

	for {
		select {
		case <-reportTick:
			workers.ReportUp("jobFetcher")

		case <-tick:
			job, err := job_repo.New().GetNext()
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

// failedJobFetcher is a worker that fetches failed jobs from the database and runs them.
func failedJobFetcher(ctx context.Context) {
	defer workers.OnStop("failedJobFetcher")

	reportTick := time.Tick(1 * time.Second)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-reportTick:
			workers.ReportUp("failedJobFetcher")

		case <-tick:
			job, err := job_repo.New().GetNextFailed()
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
