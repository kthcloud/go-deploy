package snapshot

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/models/sys/job"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/models/versions"
	"go-deploy/pkg/workers"
	"go-deploy/utils"
	"log"
	"math/rand"
	"time"
)

// snapshotter is a worker that takes snapshots.
func snapshotter(ctx context.Context) {
	defer workers.OnStop("snapshotter")

	reportTick := time.Tick(1 * time.Second)
	tick := time.Tick(10 * time.Second)

	for {
		select {
		case <-reportTick:
			workers.ReportUp("snapshotter")

		case <-tick:
			vms, err := vmModels.New().List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to get all vms. details: %w", err))
				continue
			}

			for _, vm := range vms {
				recurrings := []string{"daily", "weekly", "monthly"}

				for _, recurring := range recurrings {
					exists, err := job.New().
						IncludeTypes(job.TypeCreateSystemSnapshot).
						ExcludeStatus(job.StatusTerminated, job.StatusCompleted).
						FilterArgs("id", vm.ID).
						FilterArgs("params.name", fmt.Sprintf("auto-%s", recurring)).
						ExistsAny()
					if err != nil {
						utils.PrettyPrintError(fmt.Errorf("failed to check if snapshot job exists. details: %w", err))
						continue
					}

					if !exists {
						scheduleSnapshotJob(&vm, recurring)
					}
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func scheduleSnapshotJob(vm *vmModels.VM, recurring string) {
	log.Println("scheduling", recurring, "snapshot for vm", vm.ID)

	runAt := getRunAt(recurring)
	err := job.New().CreateScheduled(uuid.New().String(), vm.OwnerID, job.TypeCreateSystemSnapshot, versions.V1, runAt, map[string]interface{}{
		"id": vm.ID,
		"params": vmModels.CreateSnapshotParams{
			Name:        fmt.Sprintf("auto-%s", recurring),
			UserCreated: false,
			Overwrite:   true,
		},
	})

	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to create snapshot job. details: %w", err))
	}
}

func getRunAt(recurring string) time.Time {
	// randomize minutes to avoid all snapshots being created at the same time
	minutes := rand.Int() % 60

	switch recurring {
	case "daily":
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day()+1, 3, minutes, 0, 0, time.UTC)
	case "weekly":
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day()+7, 3, minutes, 0, 0, time.UTC)
	case "monthly":
		now := time.Now()
		return time.Date(now.Year(), now.Month()+1, now.Day(), 3, minutes, 0, 0, time.UTC)
	}

	log.Println("invalid recurring value:", recurring)
	return time.Time{}
}
