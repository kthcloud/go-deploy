package status_update

import (
	"context"
	"go-deploy/pkg/config"
	"go-deploy/pkg/log"
	"go-deploy/pkg/services"
)

// Setup starts the status updaters.
// Status updaters are workers that listen for status updates from external services and update the database accordingly.
func Setup(ctx context.Context) {
	log.Println("Starting status updaters")
	go services.Worker(ctx, "deploymentStatusListener", DeploymentStatusListener)
	go services.Worker(ctx, "vmStatusListener", VmStatusListener)
	go services.Worker(ctx, "eventListener", DeploymentEventListener)

	go services.PeriodicWorker(ctx, "deploymentStatusFetcher", DeploymentStatusFetcher, config.Config.Timer.DeploymentStatusUpdate)
	go services.PeriodicWorker(ctx, "vmStatusFetcher", VmStatusFetcher, config.Config.Timer.VmStatusUpdate)
	go services.PeriodicWorker(ctx, "deploymentPingUpdater", deploymentPingUpdater, config.Config.Timer.DeploymentPingUpdate)
}
