package status_update

import (
	"go-deploy/pkg/sys"
	"log"
)

func Setup(ctx *sys.Context) {
	log.Println("starting status updaters")
	go vmStatusUpdater(ctx)
	go deploymentStatusUpdater(ctx)
}
