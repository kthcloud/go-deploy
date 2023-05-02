package status_updaters

import (
	deploymentModel "go-deploy/models/deployment"
	vmModel "go-deploy/models/vm"
	"go-deploy/pkg/app"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func vmStatusUpdater(ctx *app.Context) {
	for {
		if ctx.Stop {
			break
		}

		allVms, _ := vmModel.GetAll()
		for _, vm := range allVms {
			code, message, err := fetchVmStatus(&vm)
			if err != nil {
				continue
			}
			_ = vmModel.UpdateWithBsonByID(vm.ID, bson.D{{"statusCode", code}, {"statusMessage", message}})
		}

		time.Sleep(1 * time.Second)
	}
}

func deploymentStatusUpdater(ctx *app.Context) {
	for {
		if ctx.Stop {
			break
		}

		allDeployments, _ := deploymentModel.GetAll()
		for _, deployment := range allDeployments {
			code, message, err := fetchDeploymentStatus(&deployment)
			if err != nil {
				continue
			}
			_ = deploymentModel.UpdateByID(deployment.ID, bson.D{{"statusCode", code}, {"statusMessage", message}})
		}

		time.Sleep(1 * time.Second)
	}
}
