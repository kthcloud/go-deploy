package logger

import (
	"context"
	"go-deploy/pkg/log"
	"go-deploy/pkg/workers"
)

// Setup starts the loggers.
// Loggers are used to poll logs from external services, such as Kubernetes.
// Right now, there should only be one logger running at a time.
func Setup(ctx context.Context) {
	log.Println("Starting job workers")

	go workers.Worker(ctx, "deploymentLogger", deploymentLogger)
}
