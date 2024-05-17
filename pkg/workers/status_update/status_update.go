package status_update

import (
	"context"
	"go-deploy/pkg/config"
	"go-deploy/pkg/log"
	"go-deploy/pkg/workers"
)

// Setup starts the status updaters.
// Status updaters are workers that listen for status updates from external services and update the database accordingly.
func Setup(ctx context.Context) {
	log.Println("Starting status updaters")
	go workers.Worker(ctx, "vmStatusListener", VmStatusListener)
	go workers.Worker(ctx, "deploymentStatusListener", DeploymentStatusListener)
	go workers.Worker(ctx, "eventListener", DeploymentEventListener)
	go workers.PeriodicWorker(ctx, "deploymentPingUpdater", deploymentPingUpdater, config.Config.Timer.DeploymentPingUpdate)
}
