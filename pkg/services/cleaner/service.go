package cleaner

import (
	"context"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/services"
)

// Setup starts the cleaners.
func Setup(ctx context.Context) {
	log.Println("Starting cleaners")

	go services.PeriodicWorker(ctx, "staleResourceCleaner", staleResourceCleaner, config.Config.Timer.StaleResourceCleanup)
}
