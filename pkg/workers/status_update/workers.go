package status_update

import (
	"context"
	"fmt"
	deploymentModels "go-deploy/models/sys/deployment"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/pkg/workers"
	"go-deploy/service/vm_service"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
)

func vmStatusUpdater(ctx context.Context) {
	defer workers.OnStop("vmStatusUpdater")

	for {
		select {
		case <-time.After(1 * time.Second):
			workers.ReportUp("vmStatusUpdater")

		case <-time.After(1 * time.Second):
			allVms, err := vmModels.New().List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error fetching vms: %w", err))
				continue
			}

			vsc := vm_service.New()

			for _, vm := range allVms {
				code, message, err := fetchVmStatus(&vm)
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error fetching vm status: %w", err))
					continue
				}
				_ = vmModels.New().SetWithBsonByID(vm.ID, bson.D{{"statusCode", code}, {"statusMessage", message}})

				host, err := vsc.GetHost(vm.ID)
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error fetching vm host: %w", err))
					continue
				}

				if host == nil {
					_ = vmModels.New().UpdateWithBsonByID(vm.ID, bson.D{{"$unset", bson.D{{"host", ""}}}})
				} else {
					_ = vmModels.New().SetWithBsonByID(vm.ID, bson.D{{"host", host}})
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func vmSnapshotUpdater(ctx context.Context) {
	defer workers.OnStop("vmSnapshotUpdater")

	for {
		select {
		case <-time.After(1 * time.Second):
			workers.ReportUp("vmSnapshotUpdater")

		case <-time.After(5 * time.Second):
			allVms, err := vmModels.New().List()
			if err != nil {
				log.Println("error fetching vms: ", err)
				continue
			}

			for _, vm := range allVms {
				snapshotMap := fetchSnapshotStatus(&vm)
				if snapshotMap == nil {
					continue
				}

				_ = vmModels.New().UpdateSubsystemByID(vm.ID, "cs.snapshotMap", snapshotMap)
			}
		case <-ctx.Done():
			return
		}
	}
}

func deploymentStatusUpdater(ctx context.Context) {
	defer workers.OnStop("deploymentStatusUpdater")

	for {
		select {
		case <-time.After(1 * time.Second):
			workers.ReportUp("deploymentStatusUpdater")

		case <-time.After(1 * time.Second):
			allDeployments, err := deploymentModels.New().List()
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
				_ = deploymentModels.New().SetWithBsonByID(deployment.ID, bson.D{{"statusCode", code}, {"statusMessage", message}})
			}
		case <-ctx.Done():
			return
		}
	}
}
