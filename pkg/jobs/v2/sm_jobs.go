package v2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/job_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/sm_repo"
	jErrors "github.com/kthcloud/go-deploy/pkg/jobs/errors"
	"github.com/kthcloud/go-deploy/pkg/jobs/utils"
	"github.com/kthcloud/go-deploy/service"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	"github.com/mitchellh/mapstructure"
)

func CreateSM(ctx context.Context, job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "userId", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	userID := job.Args["userId"].(string)

	var params model.SmCreateParams
	err = mapstructure.Decode(job.Args["params"].(map[string]any), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2(utils.GetAuthInfo(job)).SMs().Create(ctx, id, userID, &params)
	if err != nil {
		// We always terminate these jobs, since rerunning it would cause a ErrNonUniqueField
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func DeleteSM(ctx context.Context, job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = sm_repo.New().AddActivity(id, model.ActivityBeingDeleted)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	relatedJobs, err := job_repo.New().
		ExcludeScheduled().
		ExcludeTypes(model.JobDeleteSM).
		ExcludeStatus(model.JobStatusTerminated, model.JobStatusCompleted).
		ExcludeIDs(job.ID).
		FilterArgs("id", id).
		List()
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	if err := utils.WaitForRelatedJobs(ctx, id, relatedJobs, 300*time.Second); err != nil {
		if errors.Is(err, context.Canceled) {
			return jErrors.MakeFailedError(err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return jErrors.MakeTerminatedError(fmt.Errorf("timeout waiting for related jobs to finish: %w", err))
		}
		return jErrors.MakeFailedError(err)
	}

	err = service.V2(utils.GetAuthInfo(job)).SMs().Delete(ctx, id)
	if err != nil {
		if !errors.Is(err, sErrors.ErrSmNotFound) {
			return jErrors.MakeFailedError(err)
		}
	}

	return nil
}

func RepairSM(ctx context.Context, job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = service.V2(utils.GetAuthInfo(job)).SMs().Repair(ctx, id)
	if err != nil {
		// All errors are terminal, so we don't check for specific errors
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}
