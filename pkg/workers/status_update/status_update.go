package status_update

import (
	"context"
	"go-deploy/pkg/config"
	"go-deploy/pkg/log"
	"go-deploy/pkg/workers"
)

// Setup starts the status updaters.
// Status updaters are workers that periodically updates the status of external services, such as CloudStack.
func Setup(ctx context.Context) {
	log.Println("Starting status updaters")
	go workers.Worker(ctx, "vmStatusUpdaterV2", vmStatusUpdaterV2)
	go workers.Worker(ctx, "deploymentStatusUpdater", deploymentStatusUpdater)
	go workers.PeriodicWorker(ctx, "deploymentPingUpdater", deploymentPingUpdater, config.Config.Timer.DeploymentPingUpdate)
}
