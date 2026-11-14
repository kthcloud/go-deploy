package status_update

import (
	"context"

	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/services"
)

// Setup starts the status updaters and blocks until ctx is canceled.
// Status updaters are workers that listen for status updates from external services and update the database accordingly.
func Setup(ctx context.Context) {
	log.Println("Starting status updaters")
	services.RunWorkers(ctx,
		func(ctx context.Context) { services.Worker(ctx, "deploymentStatusListener", DeploymentStatusListener) },
		func(ctx context.Context) { services.Worker(ctx, "vmStatusListener", VmStatusListener) },
		func(ctx context.Context) { services.Worker(ctx, "eventListener", DeploymentEventListener) },
		func(ctx context.Context) { services.Worker(ctx, "gpuClaimStatusListener", GpuClaimStatusListener) },
		func(ctx context.Context) {
			services.PeriodicWorker(ctx, "deploymentStatusFetcher", DeploymentStatusFetcher, config.Config.Timer.DeploymentStatusUpdate)
		},
		func(ctx context.Context) {
			services.PeriodicWorker(ctx, "vmStatusFetcher", VmStatusFetcher, config.Config.Timer.VmStatusUpdate)
		},
		func(ctx context.Context) {
			services.PeriodicWorker(ctx, "deploymentPingUpdater", deploymentPingUpdater, config.Config.Timer.DeploymentPingUpdate)
		},
	)
}
