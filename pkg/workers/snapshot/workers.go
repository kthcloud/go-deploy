package snapshot

import (
	"fmt"
	"github.com/google/uuid"
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/db/resources/job_repo"
	"go-deploy/pkg/db/resources/vm_repo"
	"go-deploy/pkg/log"
	"go-deploy/utils"
	"math/rand"
	"time"
)

// snapshotter is a worker that takes snapshots.
func snapshotter() error {
	vms, err := vm_repo.New().List()
	if err != nil {
		return err
	}

	for _, vm := range vms {
		recurrings := []string{"daily", "weekly", "monthly"}

		for _, recurring := range recurrings {
			exists, err := job_repo.New().
				IncludeTypes(model.JobCreateSystemVmSnapshot).
				ExcludeStatus(model.JobStatusTerminated, model.JobStatusCompleted).
				FilterArgs("id", vm.ID).
				FilterArgs("params.name", fmt.Sprintf("auto-%s", recurring)).
				ExistsAny()
			if err != nil {
				return err
			}

			if !exists {
				scheduleSnapshotJob(&vm, recurring)
			}
		}
	}

	return nil
}

func scheduleSnapshotJob(vm *model.VM, recurring string) {
	runAt := getRunAt(recurring)
	err := job_repo.New().CreateScheduled(uuid.New().String(), vm.OwnerID, model.JobCreateSystemVmSnapshot, version.V1, runAt, map[string]interface{}{
		"id": vm.ID,
		"params": model.CreateSnapshotParams{
			Name:        fmt.Sprintf("auto-%s", recurring),
			UserCreated: false,
			Overwrite:   true,
		},
	})

	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to create snapshot model. details: %w", err))
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

	log.Println("Invalid recurring value:", recurring)
	return time.Time{}
}
