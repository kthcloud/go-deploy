package snapshot

import (
	"context"
	"go-deploy/pkg/log"
)

// Setup starts the snapshot workers.
// Snapshot workers are workers that periodically takes snapshots.
func Setup(ctx context.Context) {
	log.Println("Starting snapshot workers")
	// Snapshots are not implemented yet.
	//go workers.PeriodicWorker(ctx, "snapshotter", snapshotter, config.Config.Timer.Snapshot)
}
