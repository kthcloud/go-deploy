package job_execute

import (
	"context"
	"log"
)

func Setup(ctx context.Context) {
	log.Println("starting job workers")
	go jobFetcher(ctx)
	go failedJobFetcher(ctx)
}
