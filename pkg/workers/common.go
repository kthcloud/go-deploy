package workers

import (
	"go-deploy/models/sys/worker_status"
	"log"
)

func ReportStatus(name string) {
	err := worker_status.New().CreateOrUpdate(name, "running")
	if err != nil {
		log.Printf("failed to report status for worker %s. details: %s\n", name, err)
	}
}
