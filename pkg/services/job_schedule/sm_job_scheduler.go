package job_schedule

import (
	"github.com/google/uuid"
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/job_repo"
	"go-deploy/pkg/db/resources/sm_repo"
)

// SmRepairScheduler is a worker that repairs storage managers.
func SmRepairScheduler() error {
	withNoActivities, err := sm_repo.New().WithNoActivities().List()
	if err != nil {
		return err
	}

	for _, sm := range withNoActivities {
		exists, err := job_repo.New().
			IncludeTypes(model.JobRepairSM).
			ExcludeStatus(model.JobStatusTerminated, model.JobStatusCompleted).
			FilterArgs("id", sm.ID).
			ExistsAny()
		if err != nil {
			return err
		}

		if exists {
			continue
		}

		jobID := uuid.New().String()
		runAfter := GetRandomRunAfter(config.Config.Timer.SmRepair.Seconds())

		err = job_repo.New().CreateScheduled(jobID, sm.OwnerID, model.JobRepairSM, version.V2, runAfter, map[string]interface{}{
			"id": sm.ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
