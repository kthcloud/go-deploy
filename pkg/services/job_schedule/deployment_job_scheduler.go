package job_schedule

import (
	"github.com/google/uuid"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/models/version"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/db/resources/deployment_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/job_repo"
)

// DeploymentRepairScheduler is a worker that repairs deployments.
func DeploymentRepairScheduler() error {
	withNoActivities, err := deployment_repo.New().WithNoActivities().List()
	if err != nil {
		return err
	}

	for _, deployment := range withNoActivities {
		exists, err := job_repo.New().
			IncludeTypes(model.JobRepairDeployment).
			ExcludeStatus(model.JobStatusTerminated, model.JobStatusCompleted).
			FilterArgs("id", deployment.ID).
			ExistsAny()
		if err != nil {
			return err
		}

		if exists {
			continue
		}

		jobID := uuid.New().String()
		runAfter := GetRandomRunAfter(config.Config.Timer.DeploymentRepair.Seconds())

		err = job_repo.New().CreateScheduled(jobID, deployment.OwnerID, model.JobRepairDeployment, version.V2, runAfter, map[string]interface{}{
			"id": deployment.ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
