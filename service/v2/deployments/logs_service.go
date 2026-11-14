package deployments

import (
	"context"
	"fmt"
	"time"

	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/db/resources/deployment_repo"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/nats"
	"github.com/kthcloud/go-deploy/service/errors"
	"github.com/kthcloud/go-deploy/service/v2/deployments/opts"
	"github.com/kthcloud/go-deploy/utils"
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
func (c *Client) SetupLogStream(ctx context.Context, id string, handler func(string, string, string, time.Time), history int) error {
	deployment, err := c.Get(ctx, id, opts.GetOpts{Shared: true})
	if err != nil {
		return err
	}

	if deployment == nil {
		return errors.ErrDeploymentNotFound
	}

	if deployment.BeingDeleted() {
		log.Println("Deployment", id, "is being deleted. not setting up log stream")
		return nil
	}

	if nats.IsZoneLogsEnabled(deployment.Zone) {
		namespace := deploymentNamespace(deployment.Zone)
		go streamNatsLogs(ctx, deployment.Zone, namespace, deployment.Name, history, handler)
		return nil
	}

	go func() {
		handler(MessageSourceControl, "[control]", "Setting up log stream", time.Now())
		time.Sleep(500 * time.Millisecond)

		// fetch history logs
		drc := deployment_repo.New()
		drc.ResourceClient.Ctx = ctx
		logs, err := drc.GetLogs(id, history)
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
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				handler(MessageSourceKeepAlive, "[keep-alive]", "keep-alive", time.Now())
			default:
				time.Sleep(FetchPeriod)
				handler(MessageSourceControl, "[control]", "fetching logs", time.Now())

				logs, err = drc.GetLogsAfter(id, lastFetched)
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

func deploymentNamespace(zoneName string) string {
	if z := config.Config.GetZone(zoneName); z != nil && z.K8s.Namespaces.Deployment != "" {
		return z.K8s.Namespaces.Deployment
	}
	if config.Config.NATS.Logs.Namespace != "" {
		return config.Config.NATS.Logs.Namespace
	}
	return "deploy"
}

func streamNatsLogs(
	ctx context.Context,
	zoneName, namespace, deployName string,
	history int,
	handler func(string, string, string, time.Time),
) {
	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-streamCtx.Done():
				return
			case <-ticker.C:
				handler(MessageSourceKeepAlive, "[keep-alive]", "keep-alive", time.Now())
			}
		}
	}()

	if err := nats.Default.StreamLogs(streamCtx, zoneName, namespace, deployName, history, handler); err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to stream nats logs for deployment %s/%s/%s: %w", zoneName, namespace, deployName, err))
	}
}
