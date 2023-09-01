package snapshot

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/models/sys/job"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/utils"
	"log"
	"time"
)

func snapshotter(ctx context.Context) {
	defer log.Println("snapshotter stopped")

	for {
		select {
		case <-time.After(1 * time.Hour):
			vms, err := vmModel.GetAll()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to get all vms. details: %w", err))
				continue
			}

			for _, vm := range vms {
				exists, jobExistsErr := job.Exists(job.TypeCreateSnapshot, map[string]interface{}{
					"id": vm.ID,
				})

				if jobExistsErr != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to check if snapshot job exists. details: %w", jobExistsErr))
					continue
				}

				if !exists {
					log.Println("scheduling snapshotting for vm:", vm.Name)

					jobID := uuid.New().String()
					now := time.Now()
					runAt := time.Date(now.Year(), now.Month(), now.Day()+1, 3, 0, 0, 0, time.UTC)

					jobCreateErr := job.CreateScheduledJob(jobID, vm.OwnerID, job.TypeCreateSnapshot, runAt, map[string]interface{}{
						"id":          vm.ID,
						"name":        fmt.Sprintf("snapshot-%s", time.Now().Format("20060102150405")),
						"userCreated": false,
					})

					if jobCreateErr != nil {
						utils.PrettyPrintError(fmt.Errorf("failed to create snapshot job. details: %w", err))
						continue
					}
				}
			}

		case <-ctx.Done():
			return
		}
	}
}
