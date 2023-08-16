package job_execute

import (
	"go-deploy/pkg/sys"
	"log"
)

func Setup(ctx *sys.Context) {
	log.Println("starting job workers")
	go jobFetcher(ctx)
	go failedJobFetcher(ctx)
}
