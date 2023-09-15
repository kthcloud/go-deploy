package confirm

import (
	"context"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/utils"
	"log"
	"time"
)

func deploymentConfirmer(ctx context.Context) {
	defer log.Println("deploymentConfirmer stopped")

	for {
		select {
		case <-time.After(5 * time.Second):
			beingCreated, _ := deploymentModel.New().GetByActivity(deploymentModel.ActivityBeingCreated)
			for _, deployment := range beingCreated {
				created := DeploymentCreated(&deployment)
				if created {
					log.Printf("marking deployment %s as created\n", deployment.Name)
					_ = deploymentModel.New().RemoveActivity(deployment.ID, deploymentModel.ActivityBeingCreated)
				}
			}

			beingDeleted, _ := deploymentModel.New().GetByActivity(deploymentModel.ActivityBeingDeleted)
			for _, deployment := range beingDeleted {
				deleted := DeploymentDeleted(&deployment)
				if deleted {
					log.Printf("marking deployment %s as deleted\n", deployment.Name)
					_ = deploymentModel.New().DeleteByID(deployment.ID)
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func vmConfirmer(ctx context.Context) {
	defer log.Println("vmConfirmer stopped")

	for {
		select {
		case <-time.After(5 * time.Second):
			beingCreated, err := vmModel.New().GetByActivity(vmModel.ActivityBeingCreated)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to get vms being created. details: %w", err))
			}

			for _, vm := range beingCreated {
				created := VmCreated(&vm)
				if created {
					log.Printf("marking vm %s as created\n", vm.Name)
					_ = vmModel.New().RemoveActivity(vm.ID, vmModel.ActivityBeingCreated)
				}
			}

			beingDeleted, err := vmModel.New().GetByActivity(vmModel.ActivityBeingDeleted)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to get vms being deleted. details: %w", err))
			}

			for _, vm := range beingDeleted {
				deleted := VmDeleted(&vm)
				if deleted {
					log.Printf("marking vm %s as deleted\n", vm.Name)
					_ = vmModel.New().DeleteByID(vm.ID)
				}
			}

		case <-ctx.Done():
			return
		}
	}
}
