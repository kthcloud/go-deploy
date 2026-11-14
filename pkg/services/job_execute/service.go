package job_execute

import (
	"context"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/services"
)

// Setup starts the job workers and blocks until ctx is canceled.
// Job execution workers are workers that runs any jobs that are ready to be executed.
func Setup(ctx context.Context) {
	log.Println("Starting job workers")

	services.RunWorkers(ctx,
		func(ctx context.Context) {
			services.PeriodicWorkerCtx(ctx, "jobFetcher", JobFetcher, config.Config.Timer.JobFetch)
		},
		func(ctx context.Context) {
			services.PeriodicWorkerCtx(ctx, "failedJobFetcher", FailedJobFetcher, config.Config.Timer.FailedJobFetch)
		},
	)
}
