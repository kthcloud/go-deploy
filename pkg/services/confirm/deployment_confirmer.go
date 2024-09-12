package confirm

import (
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/deployment_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/job_repo"
	"github.com/kthcloud/go-deploy/pkg/log"
	"golang.org/x/exp/slices"
)

// DeploymentDeletionConfirmer is a worker that confirms deployment deletion.
// It checks if each subsystem resource is deleted, and if any related jobs are finished.
func DeploymentDeletionConfirmer() error {
	beingDeleted, err := deployment_repo.New().WithActivities(model.ActivityBeingDeleted).List()
	if err != nil {
		return err
	}

	for _, deployment := range beingDeleted {
		deleted := DeploymentDeleted(&deployment)
		if !deleted {
			continue
		}

		relatedJobs, err := job_repo.New().ExcludeScheduled().FilterArgs("id", deployment.ID).List()
		if err != nil {
			return err
		}

		allFinished := slices.IndexFunc(relatedJobs, func(j model.Job) bool {
			return j.Status != model.JobStatusCompleted &&
				j.Status != model.JobStatusTerminated
		}) == -1

		if allFinished {
			log.Printf("Marking deployment %s as deleted", deployment.ID)
			err = deployment_repo.New().DeleteByID(deployment.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
