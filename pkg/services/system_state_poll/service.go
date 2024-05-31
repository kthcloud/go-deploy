package system_state_poll

import (
	"context"
	"go-deploy/pkg/config"
	"go-deploy/pkg/log"
	"go-deploy/pkg/services"
)

func Setup(ctx context.Context) {
	log.Println("Starting pollers")

	go services.PeriodicWorker(ctx, "statsWorker", StatsWorker, config.Config.Timer.FetchSystemStats)
	go services.PeriodicWorker(ctx, "capacitiesWorker", CapacitiesWorker, config.Config.Timer.FetchSystemCapacities)
	go services.PeriodicWorker(ctx, "statusWorker", StatusWorker, config.Config.Timer.FetchSystemStatus)
	go services.PeriodicWorker(ctx, "gpuInfoWorker", GpuInfoWorker, config.Config.Timer.FetchSystemGpuInfo)
}
