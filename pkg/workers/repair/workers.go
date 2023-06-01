package repair

import (
	"github.com/google/uuid"
	deploymentModel "go-deploy/models/sys/deployment"
	jobModel "go-deploy/models/sys/job"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/app"
	"go-deploy/service/job_service"
	"log"
	"time"
)

func deploymentRepairer(ctx *app.Context) {
	firstLoop := true
	for {
		if ctx.Stop {
			break
		}

		if !firstLoop {
			time.Sleep(30 * time.Second)
		} else {
			time.Sleep(1 * time.Second)
			firstLoop = false
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
	}
}

func vmRepairer(ctx *app.Context) {
	firstLoop := true
	for {
		if ctx.Stop {
			break
		}

		if !firstLoop {
			time.Sleep(30 * time.Second)
		} else {
			time.Sleep(1 * time.Second)
			firstLoop = false
		}

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
	}
}