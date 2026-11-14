package model

import (
	body "github.com/kthcloud/go-deploy/api/openapi/v2"
	"time"
)

type WorkerStatus struct {
	Name       string    `bson:"name"`
	Status     string    `bson:"status"`
	ReportedAt time.Time `bson:"reportedAt"`
}

// ToDTO converts a WorkerStatus to a body.WorkerStatusRead DTO.
func (ws *WorkerStatus) ToDTO() body.WorkerStatusRead {
	return dtoConverter.WorkerStatusToDTO(ws)
}
