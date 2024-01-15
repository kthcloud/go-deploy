package ping

import (
	"context"
	"log"
)

// Setup starts the ping updaters.
// Ping updaters are workers that periodically pings external services, such as Kubernetes.
func Setup(ctx context.Context) {
	log.Println("starting ping updaters")
	go deploymentPingUpdater(ctx)
}
