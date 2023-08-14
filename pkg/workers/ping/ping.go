package ping

import (
	"go-deploy/pkg/sys"
	"log"
)

func Setup(ctx *sys.Context) {
	log.Println("starting status updaters")
	go deploymentPingUpdater(ctx)
}
