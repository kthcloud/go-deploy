package services

import (
	"github.com/kthcloud/go-deploy/pkg/db/resources/worker_status_repo"
	"github.com/kthcloud/go-deploy/pkg/log"
)

// ReportUp reports that a worker is up.
func ReportUp(name string) {
	err := worker_status_repo.New().CreateOrUpdate(name, "running")
	if err != nil {
		log.Printf("Failed to report status for worker %s. details: %s", name, err)
	}
}

// CleanUp deletes all worker statuses that have not been updated in the last 24 hours.
func CleanUp() {
	err := worker_status_repo.New().DeleteStale()
	if err != nil {
		log.Printf("Failed to clean up worker statuses. details: %s", err)
	}
}

// OnStop reports that a worker has stopped.
// It should be called in a defer statement for every worker.
func OnStop(name string) {
	log.Println(name, "stopped")

	err := worker_status_repo.New().CreateOrUpdate(name, "stopped")
	if err != nil {
		log.Printf("Failed to report status for worker %s. details: %s", name, err)
	}
}
