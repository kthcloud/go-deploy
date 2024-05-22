package job_execute

import (
	"context"
	"go-deploy/pkg/config"
	"go-deploy/pkg/log"
	"go-deploy/pkg/services"
)

// Setup starts the job workers.
// Job execution workers are workers that runs any jobs that are ready to be executed.
func Setup(ctx context.Context) {
	log.Println("Starting job workers")

	go services.PeriodicWorker(ctx, "jobFetcher", jobFetcher, config.Config.Timer.JobFetch)
	go services.PeriodicWorker(ctx, "failedJobFetcher", failedJobFetcher, config.Config.Timer.FailedJobFetch)
}
