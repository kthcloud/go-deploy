package confirm

import (
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/job_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/vm_repo"
	"github.com/kthcloud/go-deploy/pkg/log"
	"golang.org/x/exp/slices"
)

// VmDeletionConfirmer is a worker that confirms VM deletion.
// It checks if each subsystem resource is deleted, and if any related jobs are finished.
func VmDeletionConfirmer() error {
	beingDeleted, err := vm_repo.New().WithActivities(model.ActivityBeingDeleted).List()
	if err != nil {
		return err
	}

	for _, vm := range beingDeleted {
		deleted := VmDeleted(&vm)
		if !deleted {
			continue
		}

		relatedJobs, err := job_repo.New().ExcludeScheduled().FilterArgs("id", vm.ID).List()
		if err != nil {
			return err
		}

		allFinished := slices.IndexFunc(relatedJobs, func(j model.Job) bool {
			return j.Status != model.JobStatusCompleted &&
				j.Status != model.JobStatusTerminated
		}) == -1

		if allFinished {
			log.Printf("Marking VM %s as deleted", vm.ID)
			err = vm_repo.New().DeleteByID(vm.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
