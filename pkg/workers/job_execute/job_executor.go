package job_execute

import (
	"go-deploy/pkg/app"
	"log"
)

func Setup(ctx *app.Context) {
	log.Println("starting job workers")
	go jobFetcher(ctx)
	go failedJobFetcher(ctx)
}
