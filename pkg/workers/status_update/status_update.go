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
	go workers.PeriodicWorker(ctx, "vmStatusUpdater", vmStatusUpdater, config.Config.Timer.VmStatusUpdate)
	go workers.PeriodicWorker(ctx, "vmSnapshotUpdater", vmSnapshotUpdater, config.Config.Timer.VmSnapshotUpdate)
	go workers.PeriodicWorker(ctx, "deploymentStatusUpdater", deploymentStatusUpdater, config.Config.Timer.DeploymentStatusUpdate)
	go workers.PeriodicWorker(ctx, "deploymentPingUpdater", deploymentPingUpdater, config.Config.Timer.DeploymentPingUpdate)
}
