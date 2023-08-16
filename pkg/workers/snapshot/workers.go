package snapshot

import (
	"github.com/google/uuid"
	"go-deploy/models/sys/job"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/sys"
	"log"
	"time"
)

func snapshotter(ctx *sys.Context) {
	for {
		if ctx.Stop {
			break
		}

		time.Sleep(1 * time.Hour)

		vms, err := vmModel.GetAll()
		if err != nil {
			log.Println("failed to get all vms. details: ", err)
			continue
		}

		for _, vm := range vms {
			exists, jobExistsErr := job.Exists(job.TypeCreateSnapshot, map[string]interface{}{
				"id": vm.ID,
			})

			if jobExistsErr != nil {
				log.Println("failed to check if snapshot job exists. details: ", jobExistsErr)
				continue
			}

			if !exists {
				log.Println("scheduling snapshotting for vm:", vm.Name)

				jobID := uuid.New().String()
				now := time.Now()
				runAt := time.Date(now.Year(), now.Month(), now.Day()+1, 3, 0, 0, 0, time.UTC)

				jobCreateErr := job.CreateScheduledJob(jobID, vm.OwnerID, job.TypeCreateSnapshot, runAt, map[string]interface{}{
					"id": vm.ID,
				})

				if jobCreateErr != nil {
					log.Println("failed to create snapshot job. details: ", err)
					continue
				}
			}
		}
	}
}
