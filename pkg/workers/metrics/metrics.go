package metrics

import (
	"context"
	"log"
)

// Setup starts the metrics updaters.
// Metrics updaters are workers that periodically moves metrics into the key-value store.
func Setup(ctx context.Context) {
	log.Println("starting metrics updaters")
	go metricsWorker(ctx)
}
