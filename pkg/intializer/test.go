package intializer

import (
	deploymentModel "go-deploy/models/sys/deployment"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/service/deployment_service"
	"go-deploy/service/vm_service"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
)

func CleanUpOldTests() {
	testerID := "955f0f87-37fd-4792-90eb-9bf6989e698e"

	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Second)

	oldE2eDeployments, err := deploymentModel.GetAllWithFilter(bson.D{
		{"ownerId", testerID},
		{"createdAt", bson.D{{"$lt", oneHourAgo}}},
	})
	if err != nil {
		panic(err)
	}

	deploymentsDeleted := 0
	for _, deployment := range oldE2eDeployments {
		err = deploymentModel.ClearActivities(deployment.ID)
		if err != nil {
			panic(err)
		}

		err = deployment_service.Delete(deployment.Name)
		if err != nil {
			panic(err)
		}

		deploymentsDeleted++
	}

	oldE2eVms, err := vmModel.GetAllWithFilter(bson.D{
		{"ownerId", testerID},
		{"createdAt", bson.D{{"$lt", oneHourAgo}}},
	})

	if err != nil {
		panic(err)
	}

	vmsDeleted := 0
	for _, vm := range oldE2eVms {
		err = vmModel.ClearActivities(vm.ID)
		if err != nil {
			panic(err)
		}

		err = vm_service.Delete(vm.Name)
		if err != nil {
			panic(err)
		}

		vmsDeleted++
	}

	log.Println("cleaned up", deploymentsDeleted, "deployments and", vmsDeleted, "vms (old e2e-tests)")
}
