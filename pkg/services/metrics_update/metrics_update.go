package metrics_update

import (
	"context"
	"go-deploy/pkg/config"
	"go-deploy/pkg/log"
	"go-deploy/pkg/services"
)

// Setup starts the metrics updaters.
// Metrics updaters are workers that periodically moves metrics into the key-value store.
func Setup(ctx context.Context) {
	log.Println("Starting metrics updaters")
	go services.PeriodicWorker(ctx, "metricsUpdater", metricsUpdater, config.Config.Timer.MetricsUpdate)
}
