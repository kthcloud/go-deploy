package status_update

import (
	"context"
	"fmt"
	deploymentModels "go-deploy/models/sys/deployment"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/models/versions"
	"go-deploy/pkg/config"
	"go-deploy/pkg/workers"
	"go-deploy/service"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func vmStatusUpdater(ctx context.Context) {
	defer workers.OnStop("vmStatusUpdater")

	reportTick := time.Tick(1 * time.Second)
	tick := time.Tick(10 * time.Second)

	for {
		select {
		case <-reportTick:
			workers.ReportUp("vmStatusUpdater")

		case <-tick:
			v1Vms, err := vmModels.New(versions.V1).List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error fetching vms: %w", err))
				continue
			}

			vsc := service.V1().VMs()
			allVmStatus := make(map[string]string)

			for _, vm := range v1Vms {
				if _, ok := allVmStatus[vm.ID]; !ok {
					zone := config.Config.VM.GetZone(vm.Zone)
					if zone == nil {
						continue
					}

					statusForZone, err := vsc.CS().ListAllStatus(zone)
					if err != nil {
						utils.PrettyPrintError(fmt.Errorf("error fetching all cs vm status: %w", err))
						continue
					}

					for k, v := range statusForZone {
						allVmStatus[k] = v
					}
				}

				vmc := vmModels.New(versions.V1)

				code, message, err := fetchVmStatusV1(&vm, allVmStatus[vm.Subsystems.CS.VM.ID])
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error fetching vm status: %w", err))
					continue
				}
				_ = vmc.SetWithBsonByID(vm.ID, bson.D{{"statusCode", code}, {"statusMessage", message}})

				host, err := vsc.GetHost(vm.ID)
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error fetching vm host: %w", err))
					continue
				}

				if host == nil {
					_ = vmc.UpdateWithBsonByID(vm.ID, bson.D{{"$unset", bson.D{{"host", ""}}}})
				} else {
					_ = vmc.SetWithBsonByID(vm.ID, bson.D{{"host", host}})
				}
			}

			v2Vms, err := vmModels.New(versions.V2).List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error fetching vms: %w", err))
				continue
			}

			for _, vm := range v2Vms {
				code, message, err := fetchVmStatusV2(&vm)
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error fetching vm status: %w", err))
					continue
				}
				_ = vmModels.New(versions.V2).SetWithBsonByID(vm.ID, bson.D{{"statusCode", code}, {"statusMessage", message}})
			}

		case <-ctx.Done():
			return
		}
	}
}

func vmSnapshotUpdater(ctx context.Context) {
	defer workers.OnStop("vmSnapshotUpdater")

	reportTick := time.Tick(1 * time.Second)
	tick := time.Tick(10 * time.Second)

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
