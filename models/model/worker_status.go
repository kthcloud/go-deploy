package model

import (
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"time"
)

type WorkerStatus struct {
	Name       string    `bson:"name"`
	Status     string    `bson:"status"`
	ReportedAt time.Time `bson:"reportedAt"`
}

// ToDTO converts a WorkerStatus to a body.WorkerStatusRead DTO.
func (ws *WorkerStatus) ToDTO() body.WorkerStatusRead {
	return body.WorkerStatusRead{
		Name:       ws.Name,
		Status:     ws.Status,
		ReportedAt: ws.ReportedAt,
	}
}
