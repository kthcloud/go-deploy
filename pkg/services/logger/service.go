package logger

import (
	"context"
	"github.com/kthcloud/go-deploy/pkg/db/key_value"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/services"
	"strings"
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
	LoggerLifetime    = time.Second * 10
	LoggerUpdate      = time.Second * 5
	LoggerSynchronize = time.Second * 30
)

func LastLogKey(podName, zoneName string) string {
	return LogsKey + ":" + zoneName + ":" + podName + ":last"
}

func OwnerLogKey(podName, zoneName string) string {
	return LogsKey + ":" + zoneName + ":" + podName + ":owner"
}

func LogKey(podName, zoneName string) string {
	return LogsKey + ":" + zoneName + ":" + podName
}

func LogQueueKey(zoneName string) string {
	return "queue:" + LogsKey + ":" + zoneName
}

func PodAndZoneNameFromLogKey(key string) (podName, zoneName string) {
	// Extract logs:zone:podName
	splits := strings.Split(key, ":")
	if len(splits) > 2 {
		return splits[2], splits[1]
	}

	return "", ""
}

// ActivePods returns a list of active pods.
//
// It captures all keys that match the logs key, including:
// logs:zone:podName, logs:zone:podName:last, logs:zone:podName:owner
func ActivePods(kvc *key_value.Client, zoneName string) (map[string]bool, error) {
	keys, err := kvc.List(LogsKey + ":" + zoneName + ":*")
	if err != nil {
		return nil, err
	}

	pods := make(map[string]bool)
	for _, key := range keys {
		podName, _ := PodAndZoneNameFromLogKey(key)
		pods[podName] = true
	}

	return pods, nil
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
			go services.Worker(ctx, "deploymentLoggerControl", PodLoggerControl)
		case LogRoleWorker:
			go services.Worker(ctx, "deploymentLoggerWorker", PodLogger)
		}
	}

}
