package cleaner

import (
	"context"
	"go-deploy/pkg/config"
	"go-deploy/pkg/log"
	"go-deploy/pkg/workers"
)

// Setup starts the cleaners.
func Setup(ctx context.Context) {
	log.Println("Starting cleaners")

	go workers.PeriodicWorker(ctx, "staleResourceCleaner", staleResourceCleaner, config.Config.Timer.StaleResourceCleanup)
}
