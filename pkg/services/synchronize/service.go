package synchronize

import (
	"context"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/services"
)

// Setup starts the synchronizers.
// Synchronizers are workers that periodically synchronize resources, such as GPUs.
func Setup(ctx context.Context) {
	log.Println("Starting synchronizers")
	go services.PeriodicWorker(ctx, "gpuSynchronizer", GpuSynchronizer, config.Config.Timer.GpuSynchronize)
	go services.PeriodicWorker(ctx, "gpuLeaseSynchronizer", GpuLeaseSynchronizer, config.Config.Timer.GpuLeaseSynchronize)
}
