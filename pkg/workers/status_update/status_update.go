package status_update

import (
	"go-deploy/pkg/app"
	"log"
)

func Setup(ctx *app.Context) {
	log.Println("starting status updaters")
	go vmStatusUpdater(ctx)
	go deploymentStatusUpdater(ctx)
}
