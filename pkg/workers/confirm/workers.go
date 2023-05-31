package confirm

import (
	deploymentModel "go-deploy/models/sys/deployment"
	vmModel "go-deploy/models/sys/vm"
	gpuModel "go-deploy/models/sys/vm/gpu"
	"go-deploy/pkg/conf"

	"go-deploy/pkg/app"
	"log"
	"time"
)

func deploymentConfirmer(ctx *app.Context) {
	for {
		if ctx.Stop {
			break
		}

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
		time.Sleep(5 * time.Second)
	}
}

func vmConfirmer(ctx *app.Context) {
	for {
		if ctx.Stop {
			break
		}

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

		excludedHosts := conf.Env.GPU.ExcludedHosts

		// check if gpu lease is expired
		leased, _ := gpuModel.GetAllLeased(excludedHosts, nil)
		for _, gpu := range leased {
			if gpu.Lease.End.Before(time.Now()) {
				log.Printf("lease for gpu %s (%s) ran out, returning it...\n", gpu.ID, gpu.Data.Name)

				err := ReturnGPU(&gpu)
				if err != nil {
					log.Printf("error returning gpu %s (%s): %s\n", gpu.ID, gpu.Data.Name, err.Error())
				}
			}
		}

		time.Sleep(5 * time.Second)
	}
}
