package jobs

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

func assertParameters(job *jobModels.Job, params []string) error {
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

func waitForJob(context context.Context, job *jobModels.Job, statuses []string) error {
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

func waitForJobs(context context.Context, jobs []jobModels.Job, statuses []string) error {
	for _, job := range jobs {
		err := waitForJob(context, &job, statuses)
		if err != nil {
			return err
		}
	}

	return nil
}

func deploymentDeletedByID(id string) (bool, error) {
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

func vmDeletedByID(id string) (bool, error) {
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

// add activity to vm
func vAddActivity(activities ...string) func(*jobModels.Job) error {
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

// remove activity from vm
func vRemActivity(activities ...string) func(*jobModels.Job) error {
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

// add activity to deployment
func dAddActivity(activities ...string) func(*jobModels.Job) error {
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

// remove activity from deployment
func dRemActivity(activities ...string) func(*jobModels.Job) error {
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

// add activity to sm
func sAddActivity(activities ...string) func(*jobModels.Job) error {
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

// remove activity from sm
func sRemActivity(activities ...string) func(*jobModels.Job) error {
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

func vmDeleted(job *jobModels.Job) (bool, error) {
	id := job.Args["id"].(string)

	deleted, err := vmDeletedByID(id)
	if err != nil {
		return false, err
	}

	return deleted, nil
}

func deploymentDeleted(job *jobModels.Job) (bool, error) {
	id := job.Args["id"].(string)

	deleted, err := deploymentDeletedByID(id)
	if err != nil {
		return false, err
	}

	return deleted, nil
}

func updatingOwner(job *jobModels.Job) (bool, error) {
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

func onlyCreateSnapshotPerUser(job *jobModels.Job) (bool, error) {
	anySnapshotJob, err := jobModels.New().
		RestrictToUser(job.UserID).
		ExcludeIDs(job.ID).
		IncludeTypes(jobModels.TypeCreateUserSnapshot).
		ExcludeStatus(jobModels.StatusCompleted, jobModels.StatusTerminated).
		ExistsAny()
	if err != nil {
		return false, err
	}

	return anySnapshotJob, nil
}
