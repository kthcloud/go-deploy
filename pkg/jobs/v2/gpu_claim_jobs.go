package v2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/gpu_claim_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/job_repo"
	jErrors "github.com/kthcloud/go-deploy/pkg/jobs/errors"
	"github.com/kthcloud/go-deploy/pkg/jobs/utils"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/service"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	"github.com/mitchellh/mapstructure"
)

func CreateGpuClaim(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	var params model.GpuClaimCreateParams
	err = mapstructure.Decode(job.Args["params"].(map[string]any), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2(utils.GetAuthInfo(job)).GpuClaims().Create(id, &params)
	if err != nil {
		// We always terminate these jobs, since rerunning it would cause a ErrNonUniqueField
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func DeleteGpuClaim(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = gpu_claim_repo.New().AddActivity(id, model.ActivityBeingDeleted)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	relatedJobs, err := job_repo.New().
		ExcludeScheduled().
		ExcludeTypes(model.JobDeleteGpuClaim).
		ExcludeStatus(model.JobStatusTerminated, model.JobStatusCompleted).
		ExcludeIDs(job.ID).
		FilterArgs("id", id).
		List()
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 301*time.Second)
	defer cancel()

	done := make(chan struct{})

	go func() {
		err = utils.WaitForJobs(ctx, relatedJobs, []string{model.JobStatusCompleted, model.JobStatusTerminated})
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			if errors.Is(err, context.DeadlineExceeded) {
				log.Println("Timeout waiting for related jobs to finish for model", id)
				return
			}

			log.Println("Failed to wait for related jobs for model", id, ". details:", err)
		}
		close(done)
	}()

	select {
	case <-time.After(300 * time.Second):
		return jErrors.MakeTerminatedError(fmt.Errorf("timeout waiting for related jobs to finish"))
	case <-ctx.Done():
	case <-done:

	}

	err = service.V2(utils.GetAuthInfo(job)).SMs().Delete(id)
	if err != nil {
		if !errors.Is(err, sErrors.ErrSmNotFound) {
			return jErrors.MakeFailedError(err)
		}
	}

	return nil
}
