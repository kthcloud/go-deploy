package workers

import (
	"go-deploy/pkg/db/resources/worker_status_repo"
	"go-deploy/pkg/log"
)

// ReportUp reports that a worker is up.
func ReportUp(name string) {
	err := worker_status_repo.New().CreateOrUpdate(name, "running")
	if err != nil {
		log.Printf("failed to report status for worker %s. details: %s\n", name, err)
	}
}

// OnStop reports that a worker has stopped.
// It should be called in a defer statement for every worker.
func OnStop(name string) {
	log.Println(name, "stopped")

	err := worker_status_repo.New().CreateOrUpdate(name, "stopped")
	if err != nil {
		log.Printf("failed to report status for worker %s. details: %s\n", name, err)
	}
}
