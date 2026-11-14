package system_state_poll

import (
	"context"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/services"
)

// Setup starts the system state pollers and blocks until ctx is canceled.
func Setup(ctx context.Context) {
	log.Println("Starting pollers")

	services.RunWorkers(ctx,
		func(ctx context.Context) {
			services.PeriodicWorker(ctx, "systemStatsWorker", StatsWorker, config.Config.Timer.FetchSystemStats)
		},
		func(ctx context.Context) {
			services.PeriodicWorker(ctx, "systemCapacitiesWorker", CapacitiesWorker, config.Config.Timer.FetchSystemCapacities)
		},
		func(ctx context.Context) {
			services.PeriodicWorker(ctx, "systemStatusWorker", StatusWorker, config.Config.Timer.FetchSystemStatus)
		},
		func(ctx context.Context) {
			services.PeriodicWorker(ctx, "systemGpuInfoWorker", GpuInfoWorker, config.Config.Timer.FetchSystemGpuInfo)
		},
	)
}
