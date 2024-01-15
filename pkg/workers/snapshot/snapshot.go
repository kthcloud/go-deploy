package snapshot

import (
	"context"
	"log"
)

// Setup starts the snapshot workers.
// Snapshot workers are workers that periodically takes snapshots.
func Setup(ctx context.Context) {
	log.Println("starting snapshot workers")
	go snapshotter(ctx)
}
