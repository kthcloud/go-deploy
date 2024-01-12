package body

import "time"

type WorkerStatusRead struct {
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	ReportedAt time.Time `json:"reported_at"`
}
