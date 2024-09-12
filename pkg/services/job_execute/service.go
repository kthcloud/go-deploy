package job_execute

import (
	"context"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/services"
)

// Setup starts the job workers.
// Job execution workers are workers that runs any jobs that are ready to be executed.
func Setup(ctx context.Context) {
	log.Println("Starting job workers")

	go services.PeriodicWorker(ctx, "jobFetcher", JobFetcher, config.Config.Timer.JobFetch)
	go services.PeriodicWorker(ctx, "failedJobFetcher", FailedJobFetcher, config.Config.Timer.FailedJobFetch)
}
