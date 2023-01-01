package confirmers

import (
	deploymentModel "go-deploy/models/deployment"
	vmModel "go-deploy/models/vm"

	"go-deploy/pkg/app"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
)

func deploymentConfirmer(ctx *app.Context) {
	for {
		if ctx.Stop {
			break
		}

		beingCreated, _ := deploymentModel.GetAllDeploymentsWithFilter(bson.D{{"beingCreated", true}})
		for _, deployment := range beingCreated {
			created := DeploymentCreated(&deployment)
			if created {
				log.Printf("marking deployment %s as created\n", deployment.Name)
				_ = deploymentModel.UpdateDeployment(deployment.ID, bson.D{{"beingCreated", false}})
			}
		}

		beingDeleted, _ := deploymentModel.GetAllDeploymentsWithFilter(bson.D{{"beingDeleted", true}})
		for _, deployment := range beingDeleted {
			deleted := DeploymentDeleted(&deployment)
			if deleted {
				log.Printf("marking deployment %s as deleted\n", deployment.Name)
				_ = deploymentModel.DeleteDeployment(deployment.ID, deployment.Owner)
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

		beingCreated, _ := vmModel.GetAllWithFilter(bson.D{{"beingCreated", true}})
		for _, vm := range beingCreated {
			created := VmCreated(&vm)
			if created {
				log.Printf("marking vm %s as created\n", vm.Name)
				_ = vmModel.UpdateByID(vm.ID, bson.D{{"beingCreated", false}})
			}
		}

		beingDeleted, _ := vmModel.GetAllWithFilter(bson.D{{"beingDeleted", true}})
		for _, vm := range beingDeleted {
			deleted := VmDeleted(&vm)
			if deleted {
				log.Printf("marking vm %s as deleted\n", vm.Name)
				_ = vmModel.DeleteByID(vm.ID, vm.Owner)
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func Setup(ctx *app.Context) {
	log.Println("starting worker")
	go deploymentConfirmer(ctx)
	go vmConfirmer(ctx)
}
