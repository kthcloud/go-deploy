package confirm

import (
	"go-deploy/models/model"
	"go-deploy/pkg/db/resources/job_repo"
	"go-deploy/pkg/db/resources/sm_repo"
	"go-deploy/pkg/log"
	"golang.org/x/exp/slices"
)

// SmDeletionConfirmer is a worker that confirms SM deletion.
// It checks if each subsystem resource is deleted, and if any related jobs are finished.
func SmDeletionConfirmer() error {
	beingDeleted, err := sm_repo.New().WithActivities(model.ActivityBeingDeleted).List()
	if err != nil {
		return err
	}

	for _, sm := range beingDeleted {
		deleted := SmDeleted(&sm)
		if !deleted {
			continue
		}

		relatedJobs, err := job_repo.New().ExcludeScheduled().FilterArgs("id", sm.ID).List()
		if err != nil {
			return err
		}

		allFinished := slices.IndexFunc(relatedJobs, func(j model.Job) bool {
			return j.Status != model.JobStatusCompleted &&
				j.Status != model.JobStatusTerminated
		}) == -1

		if allFinished {
			log.Printf("Marking SM %s as deleted", sm.ID)
			err = sm_repo.New().DeleteByID(sm.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
