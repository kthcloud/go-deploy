package v2

import (
	"context"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/deployment_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/job_repo"
	jErrors "github.com/kthcloud/go-deploy/pkg/jobs/errors"
	"github.com/kthcloud/go-deploy/pkg/jobs/utils"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/services/confirm"
	"github.com/kthcloud/go-deploy/service"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	"time"
)

func CreateDeployment(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "ownerId", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	ownerID := job.Args["ownerId"].(string)
	var params body.DeploymentCreate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2(utils.GetAuthInfo(job)).Deployments().Create(id, ownerID, &params)
	if err != nil {
		var zoneCapabilityMissingErr sErrors.ZoneCapabilityMissingErr
		if errors.As(err, &zoneCapabilityMissingErr) {
			return jErrors.MakeTerminatedError(err)
		}

		// If there was some error, we trigger a repair, since rerunning it would cause a NonUniqueFieldErr
		_ = service.V2(utils.GetAuthInfo(job)).Deployments().Repair(id)
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func DeleteDeployment(job *model.Job) error {
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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

		cancel()
	}()

	select {
	case <-time.After(30 * time.Second):
		return jErrors.MakeTerminatedError(fmt.Errorf("timeout waiting for related jobs to finish"))
	case <-ctx.Done():
	}

	err = service.V2(utils.GetAuthInfo(job)).Deployments().Delete(id)
	if err != nil {
		if !errors.Is(err, sErrors.DeploymentNotFoundErr) {
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

func UpdateDeployment(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var update body.DeploymentUpdate
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &update)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2(utils.GetAuthInfo(job)).Deployments().Update(id, &update)
	if err != nil {
		switch {
		case errors.Is(err, sErrors.DeploymentNotFoundErr):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.NonUniqueFieldErr):
			return jErrors.MakeTerminatedError(err)
		case errors.Is(err, sErrors.IngressHostInUseErr):
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

func UpdateDeploymentOwner(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)
	var params model.DeploymentUpdateOwnerParams
	err = mapstructure.Decode(job.Args["params"].(map[string]interface{}), &params)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2(utils.GetAuthInfo(job)).Deployments().UpdateOwner(id, &params)
	if err != nil {
		if errors.Is(err, sErrors.DeploymentNotFoundErr) {
			return jErrors.MakeTerminatedError(err)
		}

		return jErrors.MakeFailedError(err)
	}

	if job.HasArg("resourceMigrationId") {
		resourceMigrationID := job.Args["resourceMigrationId"].(string)
		err = service.V2(utils.GetAuthInfo(job)).ResourceMigrations().Delete(resourceMigrationID)
		if err != nil {
			return jErrors.MakeTerminatedError(err)
		}
	}

	return nil
}

func RepairDeployment(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = service.V2(utils.GetAuthInfo(job)).Deployments().Repair(id)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = deployment_repo.New().MarkRepaired(id)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}
