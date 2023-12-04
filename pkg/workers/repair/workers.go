package repair

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	deploymentModel "go-deploy/models/sys/deployment"
	jobModel "go-deploy/models/sys/job"
	storageManagerModel "go-deploy/models/sys/storage_manager"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/config"
	"go-deploy/utils"
	"log"
	"math/rand"
	"time"
)

func deploymentRepairer(ctx context.Context) {
	defer log.Println("deploymentRepairer stopped")

	for {

		select {
		case <-time.After(60 * time.Second):
			restarting, err := deploymentModel.New().WithActivities(deploymentModel.ActivityRestarting).List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error fetching restarting deployments. details: %w", err))
				continue
			}

			for _, deployment := range restarting {
				// remove activity if it has been restarting for more than 5 minutes
				now := time.Now()
				if now.Sub(deployment.RestartedAt) > 5*time.Minute {
					log.Printf("removing restarting activity from deployment %s\n", deployment.Name)
					err = deploymentModel.New().RemoveActivity(deployment.ID, deploymentModel.ActivityRestarting)
					if err != nil {
						utils.PrettyPrintError(fmt.Errorf("failed to remove restarting activity from deployment %s. details: %w", deployment.Name, err))
					}
				}
			}

			withNoActivities, err := deploymentModel.New().WithNoActivities().List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error fetching deployments with no activities. details: %w", err))
				continue
			}

			for _, deployment := range withNoActivities {
				exists, err := jobModel.New().
					IncludeTypes(jobModel.TypeRepairDeployment).
					ExcludeStatus(jobModel.StatusTerminated, jobModel.StatusCompleted).
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
				// spread out repair jobs evenly over time
				seconds := config.Config.Deployment.RepairInterval + rand.Intn(config.Config.Deployment.RepairInterval)
				runAfter := time.Now().Add(time.Duration(seconds) * time.Second)

				err = jobModel.New().CreateScheduled(jobID, deployment.OwnerID, jobModel.TypeRepairDeployment, runAfter, map[string]interface{}{
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

func storageManagerRepairer(ctx context.Context) {
	defer log.Println("storageManagerRepairer stopped")

	for {
		select {
		case <-time.After(60 * time.Second):
			withNoActivities, err := storageManagerModel.New().WithNoActivities().List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error fetching storage managers with no activities. details: %w", err))
				continue
			}

			for _, storageManager := range withNoActivities {
				exists, err := jobModel.New().
					IncludeTypes(jobModel.TypeRepairStorageManager).
					ExcludeStatus(jobModel.StatusTerminated, jobModel.StatusCompleted).
					FilterArgs("id", storageManager.ID).
					ExistsAny()
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to check if repair job exists for storage manager %s. details: %w", storageManager.ID, err))
					continue
				}

				if exists {
					continue
				}

				log.Println("scheduling repair job for storage manager", storageManager.ID)

				jobID := uuid.New().String()
				// spread out repair jobs evenly over time
				seconds := config.Config.Deployment.RepairInterval + rand.Intn(config.Config.Deployment.RepairInterval)
				runAfter := time.Now().Add(time.Duration(seconds) * time.Second)

				err = jobModel.New().CreateScheduled(jobID, storageManager.OwnerID, jobModel.TypeRepairStorageManager, runAfter, map[string]interface{}{
					"id": storageManager.ID,
				})
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to schedule repair job for storage manager %s. details: %w", storageManager.ID, err))
					continue
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func vmRepairer(ctx context.Context) {
	defer log.Println("vmRepairer stopped")

	for {
		select {
		case <-time.After(60 * time.Second):
			withNoActivities, err := vmModel.New().WithNoActivities().List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error fetching vms with no activities. details: %w", err))
				continue
			}

			for _, vm := range withNoActivities {
				exists, err := jobModel.New().
					IncludeTypes(jobModel.TypeRepairVM).
					ExcludeStatus(jobModel.StatusTerminated, jobModel.StatusCompleted).
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
				// spread out repair jobs evenly over time
				seconds := config.Config.VM.RepairInterval + rand.Intn(config.Config.VM.RepairInterval)
				runAfter := time.Now().Add(time.Duration(seconds) * time.Second)

				err = jobModel.New().CreateScheduled(jobID, vm.OwnerID, jobModel.TypeRepairVM, runAfter, map[string]interface{}{
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
