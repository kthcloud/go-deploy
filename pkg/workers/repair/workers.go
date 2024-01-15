package repair

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	deploymentModels "go-deploy/models/sys/deployment"
	jobModels "go-deploy/models/sys/job"
	smModels "go-deploy/models/sys/sm"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/pkg/config"
	"go-deploy/pkg/workers"
	"go-deploy/utils"
	"log"
	"math/rand"
	"time"
)

// deploymentRepairer is a worker that repairs deployments.
func deploymentRepairer(ctx context.Context) {
	defer workers.OnStop("deploymentRepairer")

	reportTick := time.Tick(1 * time.Second)
	tick := time.Tick(time.Duration(config.Config.Deployment.RepairInterval) * time.Second)

	for {
		select {
		case <-reportTick:
			workers.ReportUp("deploymentRepairer")

		case <-tick:
			restarting, err := deploymentModels.New().WithActivities(deploymentModels.ActivityRestarting).List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error fetching restarting deployments. details: %w", err))
				continue
			}

			for _, deployment := range restarting {
				// Remove activity if it has been restarting for more than 5 minutes
				now := time.Now()
				if now.Sub(deployment.RestartedAt) > 5*time.Minute {
					log.Printf("removing restarting activity from deployment %s\n", deployment.Name)
					err = deploymentModels.New().RemoveActivity(deployment.ID, deploymentModels.ActivityRestarting)
					if err != nil {
						utils.PrettyPrintError(fmt.Errorf("failed to remove restarting activity from deployment %s. details: %w", deployment.Name, err))
					}
				}
			}

			withNoActivities, err := deploymentModels.New().WithNoActivities().List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error fetching deployments with no activities. details: %w", err))
				continue
			}

			for _, deployment := range withNoActivities {
				exists, err := jobModels.New().
					IncludeTypes(jobModels.TypeRepairDeployment).
					ExcludeStatus(jobModels.StatusTerminated, jobModels.StatusCompleted).
					FilterArgs("id", deployment.ID).
					ExistsAny()
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to check if repair job exists for deployment %s. details: %w", deployment.ID, err))
					continue
				}

				if exists {
					continue
				}

				log.Println("scheduling repair job for deployment", deployment.ID)

				jobID := uuid.New().String()
				// Spread out repair jobs evenly over time
				seconds := config.Config.Deployment.RepairInterval + rand.Intn(config.Config.Deployment.RepairInterval)
				runAfter := time.Now().Add(time.Duration(seconds) * time.Second)

				err = jobModels.New().CreateScheduled(jobID, deployment.OwnerID, jobModels.TypeRepairDeployment, runAfter, map[string]interface{}{
					"id": deployment.ID,
				})
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to schedule repair job for deployment %s. details: %w", deployment.ID, err))
					continue
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

// smRepairer is a worker that repairs storage managers.
func smRepairer(ctx context.Context) {
	defer workers.OnStop("smRepairer")

	reportTick := time.Tick(1 * time.Second)
	tick := time.Tick(time.Duration(config.Config.Deployment.RepairInterval) * time.Second)

	for {
		select {
		case <-reportTick:
			workers.ReportUp("smRepairer")

		case <-tick:
			withNoActivities, err := smModels.New().WithNoActivities().List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error fetching storage managers with no activities. details: %w", err))
				continue
			}

			for _, sm := range withNoActivities {
				exists, err := jobModels.New().
					IncludeTypes(jobModels.TypeRepairSM).
					ExcludeStatus(jobModels.StatusTerminated, jobModels.StatusCompleted).
					FilterArgs("id", sm.ID).
					ExistsAny()
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to check if repair job exists for storage manager %s. details: %w", sm.ID, err))
					continue
				}

				if exists {
					continue
				}

				log.Println("scheduling repair job for storage manager", sm.ID)

				jobID := uuid.New().String()
				// Spread out repair jobs evenly over time
				seconds := config.Config.Deployment.RepairInterval + rand.Intn(config.Config.Deployment.RepairInterval)
				runAfter := time.Now().Add(time.Duration(seconds) * time.Second)

				err = jobModels.New().CreateScheduled(jobID, sm.OwnerID, jobModels.TypeRepairSM, runAfter, map[string]interface{}{
					"id": sm.ID,
				})
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to schedule repair job for storage manager %s. details: %w", sm.ID, err))
					continue
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

// vmRepairer is a worker that repairs VMs.
func vmRepairer(ctx context.Context) {
	defer workers.OnStop("vmRepairer")

	reportTick := time.Tick(1 * time.Second)
	tick := time.Tick(time.Duration(config.Config.VM.RepairInterval) * time.Second)

	for {
		select {
		case <-reportTick:
			workers.ReportUp("vmRepairer")

		case <-tick:
			withNoActivities, err := vmModels.New().WithNoActivities().List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error fetching vms with no activities. details: %w", err))
				continue
			}

			for _, vm := range withNoActivities {
				exists, err := jobModels.New().
					IncludeTypes(jobModels.TypeRepairVM).
					ExcludeStatus(jobModels.StatusTerminated, jobModels.StatusCompleted).
					FilterArgs("id", vm.ID).
					ExistsAny()
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to check if repair job exists for vm %s. details: %w", vm.ID, err))
					continue
				}

				if exists {
					continue
				}

				log.Println("scheduling repair job for vm", vm.ID)

				jobID := uuid.New().String()
				// Spread out repair jobs evenly over time
				seconds := config.Config.VM.RepairInterval + rand.Intn(config.Config.VM.RepairInterval)
				runAfter := time.Now().Add(time.Duration(seconds) * time.Second)

				err = jobModels.New().CreateScheduled(jobID, vm.OwnerID, jobModels.TypeRepairVM, runAfter, map[string]interface{}{
					"id": vm.ID,
				})
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to schedule repair job for vm %s. details: %w", vm.ID, err))
					continue
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
