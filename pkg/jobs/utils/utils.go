package utils

import (
	"context"
	"fmt"
	deploymentModels "go-deploy/models/sys/deployment"
	jobModels "go-deploy/models/sys/job"
	smModels "go-deploy/models/sys/sm"
	vmModels "go-deploy/models/sys/vm"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/exp/slices"
	"log"
	"time"
)

// AssertParameters asserts that the job has all the required parameters.
func AssertParameters(job *jobModels.Job, params []string) error {
	for _, param := range params {
		if _, ok := job.Args[param]; !ok {
			return fmt.Errorf("missing parameter: %s", param)
		}
	}

	return nil
}

// WaitForJob waits for a job to reach one of the given statuses.
func WaitForJob(context context.Context, job *jobModels.Job, statuses []string) error {
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
			job, err = jobModels.New().GetByID(job.ID)
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
func WaitForJobs(context context.Context, jobs []jobModels.Job, statuses []string) error {
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
	deleted, err := deploymentModels.New().IncludeDeletedResources().Deleted(id)
	if err != nil {
		return false, err
	}

	if deleted {
		return true, nil
	}

	beingDeleted, err := deploymentModels.New().IsDoingActivity(id, deploymentModels.ActivityBeingDeleted)
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
	deleted, err := vmModels.New().IncludeDeletedResources().Deleted(id)
	if err != nil {
		return false, err
	}

	if deleted {
		return true, nil
	}

	beingDeleted, err := vmModels.New().IsDoingActivity(id, vmModels.ActivityBeingDeleted)
	if err != nil {
		return false, err
	}

	if beingDeleted {
		return true, nil
	}

	return false, nil
}

// VmAddActivity is a helper function that adds activity to VM
func VmAddActivity(activities ...string) func(*jobModels.Job) error {
	return func(job *jobModels.Job) error {
		id := job.Args["id"].(string)

		for _, a := range activities {
			err := vmModels.New().AddActivity(id, a)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// VmRemActivity is a helper function that removes activity from VM
func VmRemActivity(activities ...string) func(*jobModels.Job) error {
	return func(job *jobModels.Job) error {
		id := job.Args["id"].(string)

		for _, a := range activities {
			err := vmModels.New().RemoveActivity(id, a)
			if err != nil {
				return err
			}

			if a == vmModels.ActivityBeingCreated {
				log.Println("finished creating vm", id)
			}
		}
		return nil
	}
}

// DAddActivity is a helper function that adds activity to deployment
func DAddActivity(activities ...string) func(*jobModels.Job) error {
	return func(job *jobModels.Job) error {
		id := job.Args["id"].(string)

		for _, a := range activities {
			err := deploymentModels.New().AddActivity(id, a)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// DRemActivity is a helper function that removes activity from deployment
func DRemActivity(activities ...string) func(*jobModels.Job) error {
	return func(job *jobModels.Job) error {
		id := job.Args["id"].(string)

		for _, a := range activities {
			err := deploymentModels.New().RemoveActivity(id, a)
			if err != nil {
				return err
			}

			if a == deploymentModels.ActivityBeingCreated {
				log.Println("finished creating deployment", id)
			}
		}
		return nil
	}
}

// SmAddActivity is a helper function that adds activity to storage manager
func SmAddActivity(activities ...string) func(*jobModels.Job) error {
	return func(job *jobModels.Job) error {
		id := job.Args["id"].(string)

		for _, a := range activities {
			err := smModels.New().AddActivity(id, a)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// SmRemActivity is a helper function that removes activity from a storage manager
func SmRemActivity(activities ...string) func(*jobModels.Job) error {
	return func(job *jobModels.Job) error {
		id := job.Args["id"].(string)

		for _, a := range activities {
			err := smModels.New().RemoveActivity(id, a)
			if err != nil {
				return err
			}

			if a == smModels.ActivityBeingCreated {
				log.Println("finished creating sm", id)
			}
		}
		return nil
	}
}

// VmDeleted is a helper function that returns true if the VM is deleted.
func VmDeleted(job *jobModels.Job) (bool, error) {
	id := job.Args["id"].(string)

	deleted, err := VmDeletedByID(id)
	if err != nil {
		return false, err
	}

	return deleted, nil
}

// DeploymentDeleted is a helper function that returns true if the deployment is deleted.
func DeploymentDeleted(job *jobModels.Job) (bool, error) {
	id := job.Args["id"].(string)

	deleted, err := DeploymentDeletedByID(id)
	if err != nil {
		return false, err
	}

	return deleted, nil
}

// UpdatingOwner is a helper function that returns true if there is an updating owner job for the VM.
func UpdatingOwner(job *jobModels.Job) (bool, error) {
	id := job.Args["id"].(string)

	filter := bson.D{
		{"args.id", id},
		{"type", jobModels.TypeUpdateVmOwner},
		{"status", bson.D{{"$nin", []string{jobModels.StatusCompleted, jobModels.StatusTerminated}}}},
	}

	anyUpdatingOwnerJob, err := jobModels.New().AddFilter(filter).ExistsAny()
	if err != nil {
		return false, err
	}

	return anyUpdatingOwnerJob, nil
}

// OnlyCreateSnapshotPerUser is a helper function that returns true if there is a snapshot job for the user.
func OnlyCreateSnapshotPerUser(job *jobModels.Job) (bool, error) {
	anySnapshotJob, err := jobModels.New().
		RestrictToUser(job.UserID).
		ExcludeIDs(job.ID).
		IncludeTypes(jobModels.TypeCreateVmUserSnapshot).
		ExcludeStatus(jobModels.StatusCompleted, jobModels.StatusTerminated).
		ExistsAny()
	if err != nil {
		return false, err
	}

	return anySnapshotJob, nil
}
