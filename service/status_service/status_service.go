package status_service

import (
	workerStatusModels "go-deploy/models/sys/worker_status"
	"time"
)

func (c *Client) ListWorkerStatus(opts ...ListWorkerStatusOpts) ([]workerStatusModels.WorkerStatus, error) {
	workerStatus, err := workerStatusModels.New().List()
	if err != nil {
		return nil, err
	}

	for idx, status := range workerStatus {
		// If we didn't receive a status update for 10 seconds, we assume the worker is stopped
		// This is useful when the pod is being terminated, and the worker is not able to report its status
		if status.Status == "running" && time.Since(status.ReportedAt) > 10*time.Second {
			workerStatus[idx].Status = "stopped"
		}
	}

	return workerStatus, nil
}
