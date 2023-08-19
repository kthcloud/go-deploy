package repair

import (
	"context"
	"github.com/google/uuid"
	deploymentModel "go-deploy/models/sys/deployment"
	jobModel "go-deploy/models/sys/job"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/conf"
	"go-deploy/service/job_service"
	"log"
	"time"
)

func deploymentRepairer(ctx context.Context) {
	defer log.Println("deploymentRepairer stopped")

	for {

		select {
		case <-time.After(time.Duration(conf.Env.Deployment.RepairInterval) * time.Second):
			restarting, err := deploymentModel.GetByActivity(deploymentModel.ActivityRestarting)
			if err != nil {
				log.Println("error fetching restarting deployments. details: ", err)
				continue
			}

			for _, deployment := range restarting {
				// remove activity if it has been restarting for more than 5 minutes
				now := time.Now()
				if now.Sub(deployment.RestartedAt) > 5*time.Minute {
					log.Printf("removing restarting activity from deployment %s\n", deployment.Name)
					err = deploymentModel.RemoveActivity(deployment.ID, deploymentModel.ActivityRestarting)
					if err != nil {
						log.Printf("failed to remove restarting activity from deployment %s. details: %s\n", deployment.Name, err.Error())
					}
				}
			}

			withNoActivities, err := deploymentModel.GetWithNoActivities()
			if err != nil {
				log.Println("error fetching deployments with no activities. details: ", err)
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

					err = deploymentModel.MarkRepaired(deployment.ID)
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
		case <-time.After(time.Duration(conf.Env.VM.RepairInterval) * time.Second):
			withNoActivities, err := vmModel.GetWithNoActivities()
			if err != nil {
				log.Println("error fetching vms with no activities. details: ", err)
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

					err = vmModel.MarkRepaired(vm.ID)
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
		case <-time.After(time.Duration(conf.Env.GPU.RepairInterval) * time.Second):
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
