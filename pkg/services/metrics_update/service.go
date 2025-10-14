package metrics_update

import (
	"context"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/services"
)

// Setup starts the metrics updaters.
// Metrics updaters are workers that periodically moves metrics into the key-value store.
func Setup(ctx context.Context) {
	log.Println("Starting metrics updaters")
	go services.PeriodicWorker(ctx, "metricsUpdater", MetricsUpdater, config.Config.Timer.MetricsUpdate)
}
