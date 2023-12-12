package intializer

import (
	"errors"
	deploymentModel "go-deploy/models/sys/deployment"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/service/deployment_service"
	dErrors "go-deploy/service/errors"
	"go-deploy/service/vm_service"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
)

func CleanUpOldTests() {
	testerID := "955f0f87-37fd-4792-90eb-9bf6989e698e"

	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Second)

	oldE2eDeployments, err := deploymentModel.New().ListWithFilterAndProjection(bson.D{
		{"ownerId", testerID},
		{"createdAt", bson.D{{"$lt", oneHourAgo}}},
	}, nil)
	if err != nil {
		panic(err)
	}

	deploymentsDeleted := 0
	for _, deployment := range oldE2eDeployments {
		err = deploymentModel.New().ClearActivities(deployment.ID)
		if err != nil {
			panic(err)
		}

		err = deploymentModel.New().AddActivity(deployment.ID, deploymentModel.ActivityBeingDeleted)
		if err != nil {
			panic(err)
		}

		err = deployment_service.New().Delete(deployment.ID)
		if err != nil {
			if !errors.Is(err, dErrors.DeploymentNotFoundErr) {
				panic(err)
			}
		}

		deploymentsDeleted++
	}

	oldE2eVms, err := vmModel.New().ListWithFilterAndProjection(bson.D{
		{"ownerId", testerID},
		{"createdAt", bson.D{{"$lt", oneHourAgo}}},
	}, nil)

	if err != nil {
		panic(err)
	}

	vmsDeleted := 0
	for _, vm := range oldE2eVms {
		err = vmModel.New().ClearActivities(vm.ID)
		if err != nil {
			panic(err)
		}

		err = vm_service.New().Delete(vm.ID)
		if err != nil {
			panic(err)
		}

		vmsDeleted++
	}

	log.Println("cleaned up", deploymentsDeleted, "deployments and", vmsDeleted, "vms (old e2e-tests)")
}
