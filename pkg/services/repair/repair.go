package repair

import (
	"context"
	"go-deploy/pkg/config"
	"go-deploy/pkg/log"
	"go-deploy/pkg/services"
)

// Setup starts the repairers.
// Repairers are workers that periodically repairs external services, such as Kubernetes.
func Setup(ctx context.Context) {
	log.Println("Starting repairers")
	go services.PeriodicWorker(ctx, "deploymentRepairer", deploymentRepairer, config.Config.Timer.DeploymentRepair)
	go services.PeriodicWorker(ctx, "smRepairer", smRepairer, config.Config.Timer.SmRepair)
	go services.PeriodicWorker(ctx, "vmRepairer", vmRepairer, config.Config.Timer.VmRepair)
}
