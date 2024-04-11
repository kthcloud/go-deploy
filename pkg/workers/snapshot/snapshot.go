package snapshot

import (
	"context"
	"go-deploy/pkg/config"
	"go-deploy/pkg/log"
	"go-deploy/pkg/workers"
)

// Setup starts the snapshot workers.
// Snapshot workers are workers that periodically takes snapshots.
func Setup(ctx context.Context) {
	log.Println("Starting snapshot workers")
	go workers.PeriodicWorker(ctx, "snapshotter", snapshotter, config.Config.Timer.Snapshot)
}
