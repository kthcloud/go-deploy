package job_schedule

import (
	"context"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/services"
)

// Setup starts job schedulers and blocks until ctx is canceled.
func Setup(ctx context.Context) {
	log.Println("Starting job schedulers")
	services.RunWorkers(ctx,
		func(ctx context.Context) {
			services.PeriodicWorker(ctx, "deploymentRepairScheduler", DeploymentRepairScheduler, config.Config.Timer.DeploymentRepair)
		},
		func(ctx context.Context) {
			services.PeriodicWorker(ctx, "smRepairScheduler", SmRepairScheduler, config.Config.Timer.SmRepair)
		},
		func(ctx context.Context) {
			services.PeriodicWorker(ctx, "vmRepairScheduler", VmRepairScheduler, config.Config.Timer.VmRepair)
		},
	)
}
