package deployment_service

import (
	"context"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/key_value"
	"go-deploy/pkg/metrics"
	"go-deploy/service"
	"go-deploy/utils"
	"log"
	"time"
)

const (
	MessageSourceControl = "control"

	FetchPeriod = 300 * time.Millisecond
)

func SetupLogStream(ctx context.Context, id string, handler func(string, string, string), history int, auth *service.AuthInfo) error {
	deployment, err := GetByIdAuth(id, auth)
	if err != nil {
		return err
	}

	if deployment == nil {
		return DeploymentNotFoundErr
	}

	if deployment.BeingDeleted() {
		log.Println("deployment", id, "is being deleted. not setting up log stream")
		return nil
	}

	go func() {
		handler(MessageSourceControl, "[control]", "setting up log stream")
		time.Sleep(500 * time.Millisecond)

		// fetch history logs
		logs, err := deploymentModel.New().GetLogs(id, history)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to get logs for deployment %s. details: %w", id, err))
			return
		}

		for _, item := range logs {
			handler(item.Source, item.Prefix, item.Line)
		}

		// fetch live logs
		lastFetched := time.Now()
		if len(logs) > 0 {
			lastFetched = logs[len(logs)-1].CreatedAt
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(FetchPeriod)
				handler(MessageSourceControl, "[control]", "fetching logs")

				logs, err = deploymentModel.New().GetLogsAfter(id, lastFetched)
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to get logs for deployment %s after %s. details: %w", id, lastFetched, err))
					return
				}

				for _, item := range logs {
					handler(item.Source, item.Prefix, item.Line)
				}

				if len(logs) > 0 {
					lastFetched = logs[len(logs)-1].CreatedAt
				}
			}
		}
	}()

	go func() {
		err = key_value.New().Incr(metrics.KeyThreadsLog)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to increment log thread when setting up continuous log stream. details: %w", err))
		}

		for {
			time.Sleep(300 * time.Millisecond)
			if ctx.Err() != nil {
				err = key_value.New().Decr(metrics.KeyThreadsLog)
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to decrement log thread when setting up continuous log stream. details: %w", err))
				}
				return
			}
		}
	}()

	return nil
}
