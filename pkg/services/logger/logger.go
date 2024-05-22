package logger

import (
	"context"
	"go-deploy/pkg/log"
	"go-deploy/pkg/services"
	"time"
)

type LogRole string

const (
	// LogRoleControl is the role for the control logger.
	LogRoleControl = LogRole("control")
	// LogRoleWorker is the role for the worker logger.
	LogRoleWorker = LogRole("worker")

	// LogsKey is the key for logs.
	LogsKey = "logs"
)

var (
	LoggerLifetime = time.Second * 10
	LoggerUpdate   = time.Second * 5
)

func LastLogKey(podName string) string {
	return LogsKey + ":" + podName + ":last"
}

func LogKey(podName string) string {
	return LogsKey + ":" + podName
}

func LogQueueKey(zoneName string) string {
	return "queue:" + LogsKey + ":" + zoneName
}

func PodNameFromLogKey(key string) string {
	return key[len(LogsKey)+1:]
}

// Setup starts the loggers.
// Loggers are used to poll logs from external services, such as Kubernetes.
// Right now, there should only be one logger running at a time.
func Setup(ctx context.Context, roles []LogRole) {
	log.Println("Starting log worker")

	if len(roles) == 0 {
		log.Println("No loggers to start")
		return
	}

	for _, role := range roles {
		switch role {
		case LogRoleControl:
			go services.Worker(ctx, "deploymentLoggerControl", PodEventListener)
		case LogRoleWorker:
			go services.Worker(ctx, "deploymentLoggerWorker", DeploymentLogger)
		}
	}

}
