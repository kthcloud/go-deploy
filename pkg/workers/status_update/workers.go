package status_update

import (
	"context"
	"fmt"
	deploymentModels "go-deploy/models/sys/deployment"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/models/versions"
	"go-deploy/pkg/workers"
	"go-deploy/service"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func vmStatusUpdater(ctx context.Context) {
	defer workers.OnStop("vmStatusUpdater")

	reportTick := time.Tick(1 * time.Second)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-reportTick:
			workers.ReportUp("vmStatusUpdater")

		case <-tick:
			allVms, err := vmModels.New(versions.V1).List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error fetching vms: %w", err))
				continue
			}

			vsc := service.V1().VMs()

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

	reportTick := time.Tick(1 * time.Second)
	tick := time.Tick(5 * time.Second)

	for {
		select {
		case <-reportTick:
			workers.ReportUp("vmSnapshotUpdater")

		case <-tick:
			go func() {
				allVms, err := vmModels.New().List()
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error fetching vms when updating snapshot status: %w", err))
					return
				}

				for _, vm := range allVms {
					snapshotMap := fetchSnapshotStatus(&vm)
					if snapshotMap == nil {
						continue
					}

					err = vmModels.New().SetSubsystem(vm.ID, "cs.snapshotMap", snapshotMap)
					if err != nil {
						utils.PrettyPrintError(fmt.Errorf("error updating vm snapshot map: %w", err))
					}
				}
			}()
		case <-ctx.Done():
			return
		}
	}
}

func deploymentStatusUpdater(ctx context.Context) {
	defer workers.OnStop("deploymentStatusUpdater")

	reportTick := time.Tick(1 * time.Second)
	tick := time.Tick(3 * time.Second)

	for {
		select {
		case <-reportTick:
			workers.ReportUp("deploymentStatusUpdater")

		case <-tick:
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
