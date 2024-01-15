package status_update

import (
	"context"
	"log"
)

// Setup starts the status updaters.
// Status updaters are workers that periodically updates the status of external services, such as CloudStack.
func Setup(ctx context.Context) {
	log.Println("starting status updaters")
	go vmStatusUpdater(ctx)
	go vmSnapshotUpdater(ctx)
	go deploymentStatusUpdater(ctx)
}
