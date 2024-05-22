package confirm

import (
	"context"
	"go-deploy/pkg/config"
	"go-deploy/pkg/log"
	"go-deploy/pkg/services"
)

// Setup starts the confirmers.
// Confirmers are generic workers that periodically checks something until a condition is met.
func Setup(ctx context.Context) {
	log.Println("Starting confirmers")

	go services.PeriodicWorker(ctx, "deploymentDeletionConfirmer", deploymentDeletionConfirmer, config.Config.Timer.DeploymentDeletionConfirm)
	go services.PeriodicWorker(ctx, "smDeletionConfirmer", smDeletionConfirmer, config.Config.Timer.SmDeletionConfirm)
	go services.PeriodicWorker(ctx, "vmDeletionConfirmer", vmDeletionConfirmer, config.Config.Timer.VmDeletionConfirm)
	go services.PeriodicWorker(ctx, "customDomainConfirmer", customDomainConfirmer, config.Config.Timer.CustomDomainConfirm)
}
