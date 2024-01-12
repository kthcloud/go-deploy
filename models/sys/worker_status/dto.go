package worker_status

import "go-deploy/models/dto/body"

func (ws *WorkerStatus) ToDTO() body.WorkerStatusRead {
	return body.WorkerStatusRead{
		Name:       ws.Name,
		Status:     ws.Status,
		ReportedAt: ws.ReportedAt,
	}
}
