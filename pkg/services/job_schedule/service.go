package job_schedule

import (
	"context"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/services"
)

// Setup starts job schedulers
func Setup(ctx context.Context) {
	log.Println("Starting job schedulers")
	go services.PeriodicWorker(ctx, "deploymentRepairScheduler", DeploymentRepairScheduler, config.Config.Timer.DeploymentRepair)
	go services.PeriodicWorker(ctx, "smRepairScheduler", SmRepairScheduler, config.Config.Timer.SmRepair)
	go services.PeriodicWorker(ctx, "vmRepairScheduler", VmRepairScheduler, config.Config.Timer.VmRepair)
}
