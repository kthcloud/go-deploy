package repair

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	deploymentModel "go-deploy/models/sys/deployment"
	storageManagerModel "go-deploy/models/sys/deployment/storage_manager"
	jobModel "go-deploy/models/sys/job"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/config"
	"go-deploy/service/job_service"
	"go-deploy/utils"
	"log"
	"time"
)

func deploymentRepairer(ctx context.Context) {
	defer log.Println("deploymentRepairer stopped")

	for {

		select {
		case <-time.After(time.Duration(config.Config.Deployment.RepairInterval) * time.Second):
			restarting, err := deploymentModel.New().GetByActivity(deploymentModel.ActivityRestarting)
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
						log.Printf("failed to remove restarting activity from deployment %s. details: %w\n", deployment.Name, err)
					}
				}
			}

			withNoActivities, err := deploymentModel.New().GetWithNoActivities()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error fetching deployments with no activities. details: %w", err))
				continue
			}

			for _, deployment := range withNoActivities {
				now := time.Now()
				if now.Sub(deployment.RepairedAt) > 5*time.Minute {
					log.Printf("repairing deployment %s\n", deployment.Name)

					jobID := uuid.New().String()
					err = job_service.Create(jobID, deployment.OwnerID, jobModel.TypeRepairDeployment, map[string]interface{}{
						"id": deployment.ID,
					})
					if err != nil {
						log.Printf("failed to create repair job for deployment %s: %s\n", deployment.Name, err.Error())
						continue
					}

					err = deploymentModel.New().MarkRepaired(deployment.ID)
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
		case <-time.After(time.Duration(config.Config.Deployment.RepairInterval) * time.Second):
			withNoActivities, err := storageManagerModel.New().GetWithNoActivities()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error fetching storage managers with no activities. details: %w", err))
				continue
			}

			for _, storageManager := range withNoActivities {
				now := time.Now()
				if now.Sub(storageManager.RepairedAt) > 5*time.Minute {
					log.Println("repairing storage manager", storageManager.ID)

					jobID := uuid.New().String()
					err = job_service.Create(jobID, storageManager.OwnerID, jobModel.TypeRepairStorageManager, map[string]interface{}{
						"id": storageManager.ID,
					})
					if err != nil {
						utils.PrettyPrintError(fmt.Errorf("failed to create repair job for storage manager %s. details: %w", storageManager.ID, err))
						continue
					}

					err = deploymentModel.New().MarkRepaired(storageManager.ID)
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
		case <-time.After(time.Duration(config.Config.VM.RepairInterval) * time.Second):
			withNoActivities, err := vmModel.New().GetWithNoActivities()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error fetching vms with no activities. details: %w", err))
				continue
			}

			for _, vm := range withNoActivities {
				now := time.Now()
				if now.Sub(vm.RepairedAt) > 5*time.Minute {
					log.Printf("repairing vm %s\n", vm.Name)

					jobID := uuid.New().String()
					err = job_service.Create(jobID, vm.OwnerID, jobModel.TypeRepairVM, map[string]interface{}{
						"id": vm.ID,
					})
					if err != nil {
						log.Printf("failed to create repair job for vm %s: %s\n", vm.Name, err.Error())
						continue
					}

					err = vmModel.New().MarkRepaired(vm.ID)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func gpuRepairer(ctx context.Context) {
	defer log.Println("gpuRepairer stopped")

	for {
		select {
		case <-time.After(time.Duration(config.Config.GPU.RepairInterval) * time.Second):
			log.Println("repairing gpus")

			jobID := uuid.New().String()
			err := job_service.Create(jobID, "system", jobModel.TypeRepairGPUs, map[string]interface{}{})
			if err != nil {
				log.Println("failed to create repair job for gpus: ", err.Error())
				continue
			}
		case <-ctx.Done():
			return
		}
	}
}
