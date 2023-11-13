package jobs

import (
	"context"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	jobModel "go-deploy/models/sys/job"
	vmModel "go-deploy/models/sys/vm"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/exp/slices"
	"time"
)

func assertParameters(job *jobModel.Job, params []string) error {
	for _, param := range params {
		if _, ok := job.Args[param]; !ok {
			return fmt.Errorf("missing parameter: %s", param)
		}
	}

	return nil
}

func makeTerminatedError(err error) error {
	return fmt.Errorf("terminated: %w", err)
}

func makeFailedError(err error) error {
	return fmt.Errorf("failed: %w", err)
}

func waitForJob(context context.Context, job *jobModel.Job, statuses []string) error {
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
			job, err = jobModel.New().GetByID(job.ID)
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

func waitForJobs(context context.Context, jobs []jobModel.Job, statuses []string) error {
	for _, job := range jobs {
		err := waitForJob(context, &job, statuses)
		if err != nil {
			return err
		}
	}

	return nil
}

func deploymentDeletedByID(id string) (bool, error) {
	deleted, err := deploymentModel.New().IncludeDeletedResources().Deleted(id)
	if err != nil {
		return false, err
	}

	if deleted {
		return true, nil
	}

	beingDeleted, err := deploymentModel.New().DoingActivity(id, deploymentModel.ActivityBeingDeleted)
	if err != nil {
		return false, err
	}

	if beingDeleted {
		return true, nil
	}

	return false, nil
}

func vmDeletedByID(id string) (bool, error) {
	deleted, err := vmModel.New().IncludeDeletedResources().Deleted(id)
	if err != nil {
		return false, err
	}

	if deleted {
		return true, nil
	}

	beingDeleted, err := vmModel.New().DoingActivity(id, vmModel.ActivityBeingDeleted)
	if err != nil {
		return false, err
	}

	if beingDeleted {
		return true, nil
	}

	return false, nil
}

// add activity to vm
func vAddActivity(activities ...string) func(*jobModel.Job) error {
	return func(job *jobModel.Job) error {
		id := job.Args["id"].(string)

		for _, a := range activities {
			err := vmModel.New().AddActivity(id, a)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// remove activity from vm
func vRemActivity(activities ...string) func(*jobModel.Job) error {
	return func(job *jobModel.Job) error {
		id := job.Args["id"].(string)

		for _, a := range activities {
			err := vmModel.New().RemoveActivity(id, a)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// add activity to deployment
func dAddActivity(activities ...string) func(*jobModel.Job) error {
	return func(job *jobModel.Job) error {
		id := job.Args["id"].(string)

		for _, a := range activities {
			err := deploymentModel.New().AddActivity(id, a)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// remove activity from deployment
func dRemActivity(activities ...string) func(*jobModel.Job) error {
	return func(job *jobModel.Job) error {
		id := job.Args["id"].(string)

		for _, a := range activities {
			err := deploymentModel.New().RemoveActivity(id, a)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func vmDeleted(job *jobModel.Job) (bool, error) {
	id := job.Args["id"].(string)

	deleted, err := vmDeletedByID(id)
	if err != nil {
		return false, err
	}

	return deleted, nil
}

func deploymentDeleted(job *jobModel.Job) (bool, error) {
	id := job.Args["id"].(string)

	deleted, err := deploymentDeletedByID(id)
	if err != nil {
		return false, err
	}

	return deleted, nil
}

func updatingOwner(job *jobModel.Job) (bool, error) {
	id := job.Args["id"].(string)

	filter := bson.D{
		{"args.id", id},
		{"type", jobModel.TypeUpdateVmOwner},
		{"status", bson.D{{"$nin", []string{jobModel.StatusCompleted, jobModel.StatusTerminated}}}},
	}

	anyUpdatingOwnerJob, err := jobModel.New().AddFilter(filter).ExistsAny()
	if err != nil {
		return false, err
	}

	return anyUpdatingOwnerJob, nil
}

func onlyCreateSnapshotPerUser(job *jobModel.Job) (bool, error) {
	anySnapshotJob, err := jobModel.New().
		RestrictToUser(job.UserID).
		ExcludeIDs(job.ID).
		IncludeTypes(jobModel.TypeCreateUserSnapshot).
		ExcludeStatus(jobModel.StatusCompleted, jobModel.StatusTerminated).
		ExistsAny()
	if err != nil {
		return false, err
	}

	return anySnapshotJob, nil
}
