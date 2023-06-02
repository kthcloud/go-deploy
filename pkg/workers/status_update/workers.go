package status_update

import (
	deploymentModel "go-deploy/models/sys/deployment"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/sys"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
)

func vmStatusUpdater(ctx *sys.Context) {
	for {
		if ctx.Stop {
			break
		}

		time.Sleep(1 * time.Second)

		allVms, err := vmModel.GetAll()
		if err != nil {
			log.Println("error fetching vms: ", err)
			continue
		}

		for _, vm := range allVms {
			code, message, err := fetchVmStatus(&vm)
			if err != nil {
				log.Println("error fetching vm status: ", err)
				continue
			}
			_ = vmModel.UpdateWithBsonByID(vm.ID, bson.D{{"statusCode", code}, {"statusMessage", message}})
		}
	}
}

func deploymentStatusUpdater(ctx *sys.Context) {
	for {
		if ctx.Stop {
			break
		}
		time.Sleep(1 * time.Second)

		allDeployments, err := deploymentModel.GetAll()
		if err != nil {
			log.Println("error fetching deployments: ", err)
			continue
		}

		for _, deployment := range allDeployments {
			code, message, err := fetchDeploymentStatus(&deployment)
			if err != nil {
				log.Println("error fetching deployment status: ", err)
				continue
			}
			_ = deploymentModel.UpdateWithBsonByID(deployment.ID, bson.D{{"statusCode", code}, {"statusMessage", message}})
		}
	}
}
