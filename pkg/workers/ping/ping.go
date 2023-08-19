package ping

import (
	"context"
	"log"
)

func Setup(ctx context.Context) {
	log.Println("starting ping updaters")
	go deploymentPingUpdater(ctx)
}
