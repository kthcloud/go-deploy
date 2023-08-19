package confirm

import (
	"context"
	deploymentModel "go-deploy/models/sys/deployment"
	vmModel "go-deploy/models/sys/vm"
	"log"
	"time"
)

func deploymentConfirmer(ctx context.Context) {
	defer log.Println("deploymentConfirmer stopped")

	for {
		select {
		case <-time.After(5 * time.Second):
			beingCreated, _ := deploymentModel.GetByActivity(deploymentModel.ActivityBeingCreated)
			for _, deployment := range beingCreated {
				created := DeploymentCreated(&deployment)
				if created {
					log.Printf("marking deployment %s as created\n", deployment.Name)
					_ = deploymentModel.RemoveActivity(deployment.ID, deploymentModel.ActivityBeingCreated)
				}
			}

			beingDeleted, _ := deploymentModel.GetByActivity(deploymentModel.ActivityBeingDeleted)
			for _, deployment := range beingDeleted {
				deleted := DeploymentDeleted(&deployment)
				if deleted {
					log.Printf("marking deployment %s as deleted\n", deployment.Name)
					_ = deploymentModel.DeleteByID(deployment.ID, deployment.OwnerID)
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
			beingCreated, _ := vmModel.GetByActivity(vmModel.ActivityBeingCreated)
			for _, vm := range beingCreated {
				created := VmCreated(&vm)
				if created {
					log.Printf("marking vm %s as created\n", vm.Name)
					_ = vmModel.RemoveActivity(vm.ID, vmModel.ActivityBeingCreated)
				}
			}

			beingDeleted, _ := vmModel.GetByActivity(vmModel.ActivityBeingDeleted)
			for _, vm := range beingDeleted {
				deleted := VmDeleted(&vm)
				if deleted {
					log.Printf("marking vm %s as deleted\n", vm.Name)
					_ = vmModel.DeleteByID(vm.ID, vm.OwnerID)
				}
			}

		case <-ctx.Done():
			return
		}
	}
}
