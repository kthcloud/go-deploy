package status_service

import (
	statusModels "go-deploy/models/sys/worker_status"
)

func (c *Client) ListWorkerStatus(opts ...ListWorkerStatusOpts) ([]statusModels.WorkerStatus, error) {
	return statusModels.New().List()
}
