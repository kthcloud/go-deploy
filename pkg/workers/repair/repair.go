package repair

import (
	"context"
	"log"
)

// Setup starts the repairers.
// Repairers are workers that periodically repairs external services, such as Kubernetes.
func Setup(ctx context.Context) {
	log.Println("starting repairers")
	go deploymentRepairer(ctx)
	go smRepairer(ctx)
	go vmRepairer(ctx)
}
