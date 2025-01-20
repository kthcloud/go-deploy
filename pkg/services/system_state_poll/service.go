package system_state_poll

import (
	"context"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/services"
)

func Setup(ctx context.Context) {
	log.Println("Starting pollers")

	go services.PeriodicWorker(ctx, "systemStatsWorker", StatsWorker, config.Config.Timer.FetchSystemStats)
	go services.PeriodicWorker(ctx, "systemCapacitiesWorker", CapacitiesWorker, config.Config.Timer.FetchSystemCapacities)
	go services.PeriodicWorker(ctx, "systemStatusWorker", StatusWorker, config.Config.Timer.FetchSystemStatus)
	go services.PeriodicWorker(ctx, "systemGpuInfoWorker", GpuInfoWorker, config.Config.Timer.FetchSystemGpuInfo)
}
