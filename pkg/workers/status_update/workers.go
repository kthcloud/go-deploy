package status_update

import (
	"context"
	deploymentModel "go-deploy/models/sys/deployment"
	vmModel "go-deploy/models/sys/vm"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
)

func vmStatusUpdater(ctx context.Context) {
	defer log.Println("vmStatusUpdater stopped")

	for {
		select {
		case <-time.After(1 * time.Second):
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
		case <-ctx.Done():
			return
		}
	}
}

func vmSnapshotUpdater(ctx context.Context) {
	defer log.Println("vmSnapshotUpdater stopped")

	for {
		select {
		case <-time.After(5 * time.Second):
			allVms, err := vmModel.GetAll()
			if err != nil {
				log.Println("error fetching vms: ", err)
				continue
			}

			for _, vm := range allVms {
				snapshotMap := fetchSnapshotStatus(&vm)
				if snapshotMap == nil {
					continue
				}

				_ = vmModel.UpdateSubsystemByName(vm.Name, "cs", "snapshotMap", snapshotMap)
			}
		case <-ctx.Done():
			return
		}
	}
}

func deploymentStatusUpdater(ctx context.Context) {
	defer log.Println("deploymentStatusUpdater stopped")

	for {
		select {
		case <-time.After(1 * time.Second):
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
		case <-ctx.Done():
			return
		}
	}
}
