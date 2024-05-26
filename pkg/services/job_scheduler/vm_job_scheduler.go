package job_scheduler

import (
	"github.com/google/uuid"
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/job_repo"
	"go-deploy/pkg/db/resources/vm_repo"
)

// VmRepairScheduler is a worker that repairs VMs.
func VmRepairScheduler() error {
	withNoActivities, err := vm_repo.New().WithNoActivities().List()
	if err != nil {
		return err
	}

	for _, vm := range withNoActivities {
		exists, err := job_repo.New().
			IncludeTypes(model.JobRepairVM).
			ExcludeStatus(model.JobStatusTerminated, model.JobStatusCompleted).
			FilterArgs("id", vm.ID).
			ExistsAny()
		if err != nil {
			return err
		}

		if exists {
			continue
		}

		jobID := uuid.New().String()
		runAfter := GetRandomRunAfter(config.Config.Timer.VmRepair.Seconds())

		err = job_repo.New().CreateScheduled(jobID, vm.OwnerID, model.JobRepairVM, version.V2, runAfter, map[string]interface{}{
			"id": vm.ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
