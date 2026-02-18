package confirm

import (
	slices0 "slices"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/gpu_claim_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/job_repo"
	"github.com/kthcloud/go-deploy/pkg/log"
)

// GcDeletionConfirmer is a worker that confirms GC deletion.
// It checks if each subsystem resource is deleted, and if any related jobs are finished.
func GcDeletionConfirmer() error {
	beingDeleted, err := gpu_claim_repo.New().WithActivities(model.ActivityBeingDeleted).List()
	if err != nil {
		return err
	}

	for _, gc := range beingDeleted {
		deleted := GCDeleted(&gc)
		if !deleted {
			continue
		}

		relatedJobs, err := job_repo.New().ExcludeScheduled().FilterArgs("id", gc.ID).List()
		if err != nil {
			return err
		}

		allFinished := slices0.IndexFunc(([]model.Job)(relatedJobs), (func(model.Job) bool)(func(j model.Job) bool {
			return j.Status != model.JobStatusCompleted && j.Status != model.JobStatusTerminated
		})) == -1

		if allFinished {
			log.Printf("Marking GC %s as deleted", gc.ID)
			err = gpu_claim_repo.New().DeleteByID(gc.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
