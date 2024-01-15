package job_execute

import (
	"context"
	"log"
)

// Setup starts the job workers.
// Job execution workers are workers that runs any jobs that are ready to be executed.
func Setup(ctx context.Context) {
	log.Println("starting job workers")
	go jobFetcher(ctx)
	go failedJobFetcher(ctx)
}
