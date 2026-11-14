package confirm

import (
	"context"

	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/services"
)

// Setup starts the confirmers and blocks until ctx is canceled.
// Confirmers are generic workers that periodically checks something until a condition is met.
func Setup(ctx context.Context) {
	log.Println("Starting confirmers")

	services.RunWorkers(ctx,
		func(ctx context.Context) {
			services.PeriodicWorker(ctx, "deploymentDeletionConfirmer", DeploymentDeletionConfirmer, config.Config.Timer.DeploymentDeletionConfirm)
		},
		func(ctx context.Context) {
			services.PeriodicWorker(ctx, "smDeletionConfirmer", SmDeletionConfirmer, config.Config.Timer.SmDeletionConfirm)
		},
		func(ctx context.Context) {
			services.PeriodicWorker(ctx, "vmDeletionConfirmer", VmDeletionConfirmer, config.Config.Timer.VmDeletionConfirm)
		},
		func(ctx context.Context) {
			services.PeriodicWorker(ctx, "customDomainConfirmer", CustomDomainConfirmer, config.Config.Timer.CustomDomainConfirm)
		},
		func(ctx context.Context) {
			services.PeriodicWorker(ctx, "gpClaimDeletionConfirmer", GcDeletionConfirmer, config.Config.Timer.DeploymentDeletionConfirm)
		},
	)
}
