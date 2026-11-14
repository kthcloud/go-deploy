package synchronize

import (
	"context"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/services"
)

// Setup starts the synchronizers and blocks until ctx is canceled.
// Synchronizers are workers that periodically synchronize resources, such as GPUs.
func Setup(ctx context.Context) {
	log.Println("Starting synchronizers")
	services.RunWorkers(ctx,
		func(ctx context.Context) {
			services.PeriodicWorker(ctx, "gpuSynchronizer", GpuSynchronizer, config.Config.Timer.GpuSynchronize)
		},
		func(ctx context.Context) {
			services.PeriodicWorker(ctx, "gpuLeaseSynchronizer", GpuLeaseSynchronizer, config.Config.Timer.GpuLeaseSynchronize)
		},
	)
}
