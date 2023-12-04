package status_update

import (
	"context"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/service/vm_service"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
)

func vmStatusUpdater(ctx context.Context) {
	defer log.Println("vmStatusUpdater stopped")

	for {
		select {
		case <-time.After(1 * time.Second):
			allVms, err := vmModel.New().List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error fetching vms: %w", err))
				continue
			}

			for _, vm := range allVms {
				code, message, err := fetchVmStatus(&vm)
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error fetching vm status: %w", err))
					continue
				}
				_ = vmModel.New().SetWithBsonByID(vm.ID, bson.D{{"statusCode", code}, {"statusMessage", message}})

				host, err := vm_service.GetHost(vm.ID)
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error fetching vm host: %w", err))
					continue
				}

				if host == nil {
					_ = vmModel.New().UpdateWithBsonByID(vm.ID, bson.D{{"$unset", bson.D{{"host", ""}}}})
				} else {
					_ = vmModel.New().SetWithBsonByID(vm.ID, bson.D{{"host", host}})
				}
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
			allVms, err := vmModel.New().List()
			if err != nil {
				log.Println("error fetching vms: ", err)
				continue
			}

			for _, vm := range allVms {
				snapshotMap := fetchSnapshotStatus(&vm)
				if snapshotMap == nil {
					continue
				}

				_ = vmModel.New().UpdateSubsystemByID(vm.ID, "cs.snapshotMap", snapshotMap)
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
			allDeployments, err := deploymentModel.New().List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error fetching deployments. details: %w", err))
				continue
			}

			for _, deployment := range allDeployments {
				code, message, err := fetchDeploymentStatus(&deployment)
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error fetching deployment status: %w", err))
					continue
				}
				_ = deploymentModel.New().SetWithBsonByID(deployment.ID, bson.D{{"statusCode", code}, {"statusMessage", message}})
			}
		case <-ctx.Done():
			return
		}
	}
}
