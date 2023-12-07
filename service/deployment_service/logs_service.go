package deployment_service

import (
	"context"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/service/deployment_service/client"
	"go-deploy/service/deployment_service/errors"
	"go-deploy/utils"
	"log"
	"time"
)

const (
	MessageSourceControl = "control"

	FetchPeriod = 300 * time.Millisecond
)

func (c *Client) SetupLogStream(ctx context.Context, handler func(string, string, string), history int) error {
	deployment, err := c.Get(&client.GetOptions{Shared: true})
	if err != nil {
		return err
	}

	if deployment == nil {
		return errors.DeploymentNotFoundErr
	}

	if deployment.BeingDeleted() {
		log.Println("deployment", c.ID, "is being deleted. not setting up log stream")
		return nil
	}

	go func() {
		handler(MessageSourceControl, "[control]", "setting up log stream")
		time.Sleep(500 * time.Millisecond)

		// fetch history logs
		logs, err := deploymentModel.New().GetLogs(c.ID(), history)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to get logs for deployment %s. details: %w", c.ID, err))
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

				logs, err = deploymentModel.New().GetLogsAfter(c.ID(), lastFetched)
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to get logs for deployment %s after %s. details: %w", c.ID, lastFetched, err))
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

	return nil
}
