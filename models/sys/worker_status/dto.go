package worker_status

import "go-deploy/models/dto/body"

// ToDTO converts a WorkerStatus to a body.WorkerStatusRead DTO.
func (ws *WorkerStatus) ToDTO() body.WorkerStatusRead {
	return body.WorkerStatusRead{
		Name:       ws.Name,
		Status:     ws.Status,
		ReportedAt: ws.ReportedAt,
	}
}
