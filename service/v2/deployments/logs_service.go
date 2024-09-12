package deployments

import (
	"context"
	"fmt"
	"github.com/kthcloud/go-deploy/pkg/db/resources/deployment_repo"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/service/errors"
	"github.com/kthcloud/go-deploy/service/v2/deployments/opts"
	"github.com/kthcloud/go-deploy/utils"
	"time"
)

const (
	// MessageSourceControl is the source of the message from the control.
	// The handler should ignore these messages, as they are only intended to check if the log stream is working.
	MessageSourceControl = "control"

	// MessageSourceKeepAlive is the source of the message from the keep alive.
	// They are sent to the client to keep the connection alive.
	MessageSourceKeepAlive = "keep-alive"

	// FetchPeriod is the period between each fetch of logs from the database.
	FetchPeriod = 300 * time.Millisecond
)

// SetupLogStream sets up a log stream for the deployment.
//
// It will continuously check the deployment logs and read the logs after the last read log.
// Increasing the history will increase the time it takes to set up the log stream.
func (c *Client) SetupLogStream(id string, ctx context.Context, handler func(string, string, string, time.Time), history int) error {
	deployment, err := c.Get(id, opts.GetOpts{Shared: true})
	if err != nil {
		return err
	}

	if deployment == nil {
		return errors.DeploymentNotFoundErr
	}

	if deployment.BeingDeleted() {
		log.Println("Deployment", id, "is being deleted. not setting up log stream")
		return nil
	}

	go func() {
		handler(MessageSourceControl, "[control]", "Setting up log stream", time.Now())
		time.Sleep(500 * time.Millisecond)

		// fetch history logs
		logs, err := deployment_repo.New().GetLogs(id, history)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to get logs for deployment %s. details: %w", id, err))
			return
		}

		for _, item := range logs {
			handler(item.Source, item.Prefix, item.Line, item.CreatedAt)
		}

		// Fetch live logs
		lastFetched := time.Now()
		if len(logs) > 0 {
			lastFetched = logs[len(logs)-1].CreatedAt
		}

		// Keep-alive packet every 30 seconds
		ticker := time.Tick(30 * time.Second)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker:
				handler(MessageSourceKeepAlive, "[keep-alive]", "keep-alive", time.Now())
			default:
				time.Sleep(FetchPeriod)
				handler(MessageSourceControl, "[control]", "fetching logs", time.Now())

				logs, err = deployment_repo.New().GetLogsAfter(id, lastFetched)
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to get logs for deployment %s after %s. details: %w", id, lastFetched, err))
					return
				}

				for _, item := range logs {
					handler(item.Source, item.Prefix, item.Line, item.CreatedAt)
				}

				if len(logs) > 0 {
					lastFetched = logs[len(logs)-1].CreatedAt
				}
			}
		}
	}()

	return nil
}
