package logger

import (
	"context"
	"log"
)

// Setup starts the loggers.
// Loggers are used to poll logs from external services, such as Kubernetes.
// Right now, there should only be one logger running at a time.
func Setup(ctx context.Context) {
	log.Println("starting job workers")
	go deploymentLogger(ctx)
}
