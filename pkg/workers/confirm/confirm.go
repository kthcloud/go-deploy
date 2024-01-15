package confirm

import (
	"context"
	"log"
)

// Setup starts the confirmers.
// Confirmers are generic workers that periodically checks something until a condition is met.
func Setup(ctx context.Context) {
	log.Println("starting confirmers")
	go deploymentDeletionConfirmer(ctx)
	go customDomainConfirmer(ctx)
	go smDeletionConfirmer(ctx)
	go vmDeletionConfirmer(ctx)
}
