package worker_status

import "time"

type WorkerStatus struct {
	Name       string    `bson:"name"`
	Status     string    `bson:"status"`
	ReportedAt time.Time `bson:"reportedAt"`
}
