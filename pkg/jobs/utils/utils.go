package utils

import (
	"context"
	"fmt"
	"go-deploy/models/model"
	"go-deploy/pkg/db/resources/deployment_repo"
	"go-deploy/pkg/db/resources/job_repo"
	"go-deploy/pkg/db/resources/vm_repo"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/exp/slices"
	"log"
	"time"
)

// AssertParameters asserts that the job has all the required parameters.
func AssertParameters(job *model.Job, params []string) error {
	for _, param := range params {
		if _, ok := job.Args[param]; !ok {
			return fmt.Errorf("missing parameter: %s", param)
		}
	}

	return nil
}

// WaitForJob waits for a job to reach one of the given statuses.
func WaitForJob(context context.Context, job *model.Job, statuses []string) error {
	if len(statuses) == 0 {
		return nil
	}

	if slices.IndexFunc(statuses, func(s string) bool { return s == job.Status }) != -1 {
		return nil
	}

	for {
		select {
		case <-context.Done():
			return context.Err()
		default:
			var err error
			job, err = job_repo.New().GetByID(job.ID)
			if err != nil {
				return err
			}

			if slices.IndexFunc(statuses, func(s string) bool { return s == job.Status }) != -1 {
				return nil
			}

			time.Sleep(1 * time.Second)
		}
	}
}

// WaitForJobs waits for a list of jobs to reach one of the given statuses.
func WaitForJobs(context context.Context, jobs []model.Job, statuses []string) error {
	for _, job := range jobs {
		err := WaitForJob(context, &job, statuses)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeploymentDeletedByID returns true if the deployment is deleted.
func DeploymentDeletedByID(id string) (bool, error) {
	deleted, err := deployment_repo.New().IncludeDeletedResources().Deleted(id)
	if err != nil {
		return false, err
	}

	if deleted {
		return true, nil
	}

	beingDeleted, err := deployment_repo.New().IsDoingActivity(id, model.ActivityBeingDeleted)
	if err != nil {
		return false, err
	}

	if beingDeleted {
		return true, nil
	}

	return false, nil
}

// VmDeletedByID returns true if the VM is deleted.
func VmDeletedByID(id string) (bool, error) {
	deleted, err := vm_repo.New().IncludeDeletedResources().Deleted(id)
	if err != nil {
		return false, err
	}

	if deleted {
		return true, nil
	}

	beingDeleted, err := vm_repo.New().IsDoingActivity(id, model.ActivityBeingDeleted)
	if err != nil {
		return false, err
	}

	if beingDeleted {
		return true, nil
	}

	return false, nil
}

// VmAddActivity is a helper function that adds activity to VM
func VmAddActivity(activities ...string) func(*model.Job) error {
	return func(job *model.Job) error {
		id := job.Args["id"].(string)

		for _, a := range activities {
			err := vm_repo.New().AddActivity(id, a)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// VmRemActivity is a helper function that removes activity from VM
func VmRemActivity(activities ...string) func(*model.Job) error {
	return func(job *model.Job) error {
		id := job.Args["id"].(string)

		for _, a := range activities {
			err := vm_repo.New().RemoveActivity(id, a)
			if err != nil {
				return err
			}

			if a == model.ActivityBeingCreated {
				log.Println("finished creating vm", id)
			}
		}
		return nil
	}
}

// DAddActivity is a helper function that adds activity to deployment
func DAddActivity(activities ...string) func(*model.Job) error {
	return func(job *model.Job) error {
		id := job.Args["id"].(string)

		for _, a := range activities {
			err := deployment_repo.New().AddActivity(id, a)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// DRemActivity is a helper function that removes activity from deployment
func DRemActivity(activities ...string) func(*model.Job) error {
	return func(job *model.Job) error {
		id := job.Args["id"].(string)

		for _, a := range activities {
			err := deployment_repo.New().RemoveActivity(id, a)
			if err != nil {
				return err
			}

			if a == model.ActivityBeingCreated {
				log.Println("finished creating deployment", id)
			}
		}
		return nil
	}
}

// VmDeleted is a helper function that returns true if the VM is deleted.
func VmDeleted(job *model.Job) (bool, error) {
	id := job.Args["id"].(string)

	deleted, err := VmDeletedByID(id)
	if err != nil {
		return false, err
	}

	return deleted, nil
}

// DeploymentDeleted is a helper function that returns true if the deployment is deleted.
func DeploymentDeleted(job *model.Job) (bool, error) {
	id := job.Args["id"].(string)

	deleted, err := DeploymentDeletedByID(id)
	if err != nil {
		return false, err
	}

	return deleted, nil
}

// UpdatingOwner is a helper function that returns true if there is an updating owner job for the VM.
func UpdatingOwner(job *model.Job) (bool, error) {
	id := job.Args["id"].(string)

	filter := bson.D{
		{"args.id", id},
		{"type", model.JobUpdateVmOwner},
		{"status", bson.D{{"$nin", []string{model.JobStatusCompleted, model.JobStatusTerminated}}}},
	}

	anyUpdatingOwnerJob, err := job_repo.New().AddFilter(filter).ExistsAny()
	if err != nil {
		return false, err
	}

	return anyUpdatingOwnerJob, nil
}

// OnlyCreateSnapshotPerUser is a helper function that returns true if there is a snapshot job for the user.
func OnlyCreateSnapshotPerUser(job *model.Job) (bool, error) {
	anySnapshotJob, err := job_repo.New().
		RestrictToUser(job.UserID).
		ExcludeIDs(job.ID).
		IncludeTypes(model.JobCreateVmUserSnapshot).
		ExcludeStatus(model.JobStatusCompleted, model.JobStatusTerminated).
		ExistsAny()
	if err != nil {
		return false, err
	}

	return anySnapshotJob, nil
}
