package repair

import (
	"github.com/google/uuid"
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/deployment_repo"
	"go-deploy/pkg/db/resources/job_repo"
	"go-deploy/pkg/db/resources/sm_repo"
	"go-deploy/pkg/db/resources/vm_repo"
	"go-deploy/pkg/log"
	"math/rand"
	"time"
)

// deploymentRepairer is a worker that repairs deployments.
func deploymentRepairer() error {
	restarting, err := deployment_repo.New().WithActivities(model.ActivityRestarting).List()
	if err != nil {
		return err
	}

	for _, deployment := range restarting {
		// Remove activity if it has been restarting for more than 5 minutes
		now := time.Now()
		if now.Sub(deployment.RestartedAt) > 5*time.Minute {
			log.Printf("removing restarting activity from deployment %s\n", deployment.Name)
			err = deployment_repo.New().RemoveActivity(deployment.ID, model.ActivityRestarting)
			if err != nil {
				return err
			}
		}
	}

	withNoActivities, err := deployment_repo.New().WithNoActivities().List()
	if err != nil {
		return err
	}

	for _, deployment := range withNoActivities {
		exists, err := job_repo.New().
			IncludeTypes(model.JobRepairDeployment).
			ExcludeStatus(model.JobStatusTerminated, model.JobStatusCompleted).
			FilterArgs("id", deployment.ID).
			ExistsAny()
		if err != nil {
			return err
		}

		if exists {
			continue
		}

		jobID := uuid.New().String()
		// Spread out repair jobs evenly over time
		seconds := int(config.Config.Timer.DeploymentRepair.Seconds()) + rand.Intn(int(config.Config.Timer.DeploymentRepair.Seconds()))
		runAfter := time.Now().Add(time.Duration(seconds) * time.Second)

		err = job_repo.New().CreateScheduled(jobID, deployment.OwnerID, model.JobRepairDeployment, version.V1, runAfter, map[string]interface{}{
			"id": deployment.ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// smRepairer is a worker that repairs storage managers.
func smRepairer() error {
	withNoActivities, err := sm_repo.New().WithNoActivities().List()
	if err != nil {
		return err
	}

	for _, sm := range withNoActivities {
		exists, err := job_repo.New().
			IncludeTypes(model.JobRepairSM).
			ExcludeStatus(model.JobStatusTerminated, model.JobStatusCompleted).
			FilterArgs("id", sm.ID).
			ExistsAny()
		if err != nil {
			return err
		}

		if exists {
			continue
		}

		jobID := uuid.New().String()
		// Spread out repair jobs evenly over time
		seconds := int(config.Config.Timer.SmRepair.Seconds()) + rand.Intn(int(config.Config.Timer.SmRepair.Seconds()))
		runAfter := time.Now().Add(time.Duration(seconds) * time.Second)

		err = job_repo.New().CreateScheduled(jobID, sm.OwnerID, model.JobRepairSM, version.V1, runAfter, map[string]interface{}{
			"id": sm.ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// vmRepairer is a worker that repairs VMs.
func vmRepairer() error {
	withNoActivities, err := vm_repo.New().WithNoActivities().List()
	if err != nil {
		return err
	}

	for _, vm := range withNoActivities {
		exists, err := job_repo.New().
			IncludeTypes(model.JobRepairVM).
			ExcludeStatus(model.JobStatusTerminated, model.JobStatusCompleted).
			FilterArgs("id", vm.ID).
			ExistsAny()
		if err != nil {
			return err
		}

		if exists {
			continue
		}

		jobID := uuid.New().String()
		// Spread out repair jobs evenly over time
		seconds := int(config.Config.Timer.VmRepair.Seconds()) + rand.Intn(int(config.Config.Timer.VmRepair.Seconds()))
		runAfter := time.Now().Add(time.Duration(seconds) * time.Second)

		err = job_repo.New().CreateScheduled(jobID, vm.OwnerID, model.JobRepairVM, version.V1, runAfter, map[string]interface{}{
			"id": vm.ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
