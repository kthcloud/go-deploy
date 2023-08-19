package status_update

import (
	"context"
	"log"
)

func Setup(ctx context.Context) {
	log.Println("starting status updaters")
	go vmStatusUpdater(ctx)
	go vmSnapshotUpdater(ctx)
	go deploymentStatusUpdater(ctx)
}
