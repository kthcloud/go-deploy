package synchronize

import (
	"context"
	"go-deploy/pkg/config"
	"go-deploy/pkg/log"
	"go-deploy/pkg/workers"
)

// Setup starts the synchronizers.
// Synchronizers are workers that periodically synchronize resources, such as GPUs.
func Setup(ctx context.Context) {
	log.Println("Starting synchronizers")
	go workers.PeriodicWorker(ctx, "gpuSynchronizer", GpuSynchronizer, config.Config.Timer.GpuSynchronize)
	go workers.PeriodicWorker(ctx, "gpuLeaseSynchronizer", GpuLeaseSynchronizer, config.Config.Timer.GpuLeaseSynchronize)
}
