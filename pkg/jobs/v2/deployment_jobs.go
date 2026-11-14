package v2

import (
	"context"
	"errors"
	"fmt"
	"time"

	body "github.com/kthcloud/go-deploy/api/openapi/v2"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/deployment_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/job_repo"
	jErrors "github.com/kthcloud/go-deploy/pkg/jobs/errors"
	"github.com/kthcloud/go-deploy/pkg/jobs/utils"
	"github.com/kthcloud/go-deploy/pkg/services/confirm"
	"github.com/kthcloud/go-deploy/service"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	"github.com/mitchellh/mapstructure"
)

func CreateDeployment(ctx context.Context, job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "ownerId", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	ownerID := job.Args["ownerId"].(string)
	var params body.DeploymentCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]any), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2(utils.GetAuthInfo(job)).Deployments().Create(ctx, id, ownerID, &params)
	if err != nil {
		if _, ok := errors.AsType[sErrors.ZoneCapabilityMissingErr](err); ok {
			return jErrors.MakeTerminatedError(err)
		}

		// If there was some error, we trigger a repair, since rerunning it would cause a ErrNonUniqueField
		_ = service.V2(utils.GetAuthInfo(job)).Deployments().Repair(ctx, id)
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func DeleteDeployment(ctx context.Context, job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = deployment_repo.New().AddActivity(id, model.ActivityBeingDeleted)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	relatedJobs, err := job_repo.New().
		ExcludeScheduled().
		ExcludeTypes(model.JobDeleteDeployment).
		ExcludeStatus(model.JobStatusTerminated, model.JobStatusCompleted).
		ExcludeIDs(job.ID).
		FilterArgs("id", id).
		List()
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	if err := utils.WaitForRelatedJobs(ctx, id, relatedJobs, 30*time.Second); err != nil {
		if errors.Is(err, context.Canceled) {
			return jErrors.MakeFailedError(err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return jErrors.MakeTerminatedError(fmt.Errorf("timeout waiting for related jobs to finish: %w", err))
		}
		return jErrors.MakeFailedError(err)
	}

	err = service.V2(utils.GetAuthInfo(job)).Deployments().Delete(ctx, id)
	if err != nil {
		if !errors.Is(err, sErrors.ErrDeploymentNotFound) {
			return jErrors.MakeFailedError(err)
		}
	}

	// Check if deleted, otherwise mark as failed and return to queue for retry
	deployment, err := deployment_repo.New().GetByID(id)
	if err != nil {
		return jErrors.MakeFailedError(err)
	}

	if deployment != nil {
		if confirm.DeploymentDeleted(deployment) {
			return nil
		}

		return jErrors.MakeFailedError(fmt.Errorf("deployment not deleted"))
	}

	return nil
}

func UpdateDeployment(ctx context.Context, job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var update body.DeploymentUpdate
	err = mapstructure.Decode(job.Args["params"].(map[string]any), &update)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2(utils.GetAuthInfo(job)).Deployments().Update(ctx, id, &update)
	if err != nil {
		switch {
		case errors.Is(err, sErrors.ErrDeploymentNotFound):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.ErrNonUniqueField):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.ErrIngressHostInUse):
			return jErrors.MakeTerminatedError(err)
		}

		return jErrors.MakeFailedError(err)
	}

	err = deployment_repo.New().MarkUpdated(id)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func UpdateDeploymentOwner(ctx context.Context, job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var params model.DeploymentUpdateOwnerParams
	err = mapstructure.Decode(job.Args["params"].(map[string]any), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2(utils.GetAuthInfo(job)).Deployments().UpdateOwner(ctx, id, &params)
	if err != nil {
		if errors.Is(err, sErrors.ErrDeploymentNotFound) {
			return jErrors.MakeTerminatedError(err)
		}

		return jErrors.MakeFailedError(err)
	}

	if job.HasArg("resourceMigrationId") {
		resourceMigrationID := job.Args["resourceMigrationId"].(string)
		err = service.V2(utils.GetAuthInfo(job)).ResourceMigrations().Delete(ctx, resourceMigrationID)
		if err != nil {
			return jErrors.MakeTerminatedError(err)
		}
	}

	return nil
}

func RepairDeployment(ctx context.Context, job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = service.V2(utils.GetAuthInfo(job)).Deployments().Repair(ctx, id)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = deployment_repo.New().MarkRepaired(id)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}
