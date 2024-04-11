package repair

import (
	"context"
	"go-deploy/pkg/config"
	"go-deploy/pkg/log"
	"go-deploy/pkg/workers"
)

// Setup starts the repairers.
// Repairers are workers that periodically repairs external services, such as Kubernetes.
func Setup(ctx context.Context) {
	log.Println("Starting repairers")
	go workers.PeriodicWorker(ctx, "deploymentRepairer", deploymentRepairer, config.Config.Timer.DeploymentRepair)
	go workers.PeriodicWorker(ctx, "smRepairer", smRepairer, config.Config.Timer.SmRepair)
	go workers.PeriodicWorker(ctx, "vmRepairer", vmRepairer, config.Config.Timer.VmRepair)
}
