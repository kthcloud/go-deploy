package utils

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/deployment_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/job_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/vm_repo"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/service/core"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"slices"
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

// GetAuthInfo returns the auth info from the job.
// AuthInfo is not always available in the job, so it might be nil.
func GetAuthInfo(job *model.Job) *core.AuthInfo {
	if job.Args == nil {
		return nil
	}

	if job.Args["authInfo"] == nil {
		return nil
	}

	toTimeHookFunc := func() mapstructure.DecodeHookFunc {
		return func(f reflect.Type, t reflect.Type, data any) (any, error) {
			if t != reflect.TypeFor[time.Time]() {
				return data, nil
			}

			switch f.Kind() {
			case reflect.Int64:
				return time.Unix(0, int64(data.(primitive.DateTime))*int64(time.Millisecond)), nil
			default:
				return data, nil
			}
		}
	}

	authInfo := &core.AuthInfo{}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			toTimeHookFunc()),
		Result: &authInfo,
	})
	if err != nil {
		return nil
	}

	err = decoder.Decode(job.Args["authInfo"].(map[string]any))
	if err != nil {
		return nil
	}

	return authInfo
}

// WaitForJob waits for a job to reach one of the given statuses.
func WaitForJob(ctx context.Context, job *model.Job, statuses []string) error {
	if len(statuses) == 0 || job == nil {
		return nil
	}

	if slices.Contains(statuses, job.Status) {
		return nil
	}

	repo := job_repo.New()
	repo.Ctx = ctx

	for {
		timer := time.NewTimer(1 * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			var err error
			job, err = repo.GetByID(job.ID)
			if err != nil {
				return err
			}
			if job == nil {
				return nil
			}

			if slices.Contains(statuses, job.Status) {
				return nil
			}
		}
	}
}

// WaitForJobs waits for a list of jobs to reach one of the given statuses.
func WaitForJobs(ctx context.Context, jobs []model.Job, statuses []string) error {
	for _, job := range jobs {
		if err := WaitForJob(ctx, &job, statuses); err != nil {
			return err
		}
	}

	return nil
}

// WaitForRelatedJobs waits up to timeout for related jobs to reach a terminal status.
func WaitForRelatedJobs(ctx context.Context, modelID string, relatedJobs []model.Job, timeout time.Duration) error {
	if len(relatedJobs) == 0 {
		return nil
	}

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := WaitForJobs(waitCtx, relatedJobs, []string{model.JobStatusCompleted, model.JobStatusTerminated})
	if err != nil {
		if waitCtx.Err() == context.DeadlineExceeded {
			log.Println("Timeout waiting for related jobs to finish for model", modelID)
		}
		return err
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
				log.Println("Finished creating vm", id)
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
				log.Println("Finished creating deployment", id)
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
		{Key: "args.id", Value: id},
		{Key: "type", Value: model.JobUpdateVmOwner},
		{Key: "status", Value: bson.D{{Key: "$nin", Value: []string{model.JobStatusCompleted, model.JobStatusTerminated}}}},
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
		WithUserID(job.UserID).
		ExcludeIDs(job.ID).
		IncludeTypes(model.JobCreateVmUserSnapshot).
		ExcludeStatus(model.JobStatusCompleted, model.JobStatusTerminated).
		ExistsAny()
	if err != nil {
		return false, err
	}

	return anySnapshotJob, nil
}
