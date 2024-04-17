package status_update

import (
	"fmt"
	"go-deploy/models/version"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/deployment_repo"
	"go-deploy/pkg/db/resources/vm_repo"
	"go-deploy/service"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
)

func vmStatusUpdater() error {
	v1Vms, err := vm_repo.New(version.V1).List()
	if err != nil {
		return err
	}

	vsc := service.V1().VMs()
	allVmStatus := make(map[string]string)

	for _, vm := range v1Vms {
		if _, ok := allVmStatus[vm.ID]; !ok {
			zone := config.Config.VM.GetLegacyZone(vm.Zone)
			if zone == nil {
				continue
			}

			statusForZone, err := vsc.CS().ListAllStatus(zone)
			if err != nil {
				return err
			}

			for k, v := range statusForZone {
				allVmStatus[k] = v
			}
		}

		vmc := vm_repo.New(version.V1)

		code, message, err := fetchVmStatusV1(&vm, allVmStatus[vm.Subsystems.CS.VM.ID])
		if err != nil {
			return err
		}
		err = vmc.SetWithBsonByID(vm.ID, bson.D{{"statusCode", code}, {"statusMessage", message}})
		if err != nil {
			return err
		}

		host, err := vsc.GetHost(vm.ID)
		if err != nil {
			return err
		}

		if host == nil {
			err = vmc.UpdateWithBsonByID(vm.ID, bson.D{{"$unset", bson.D{{"host", ""}}}})
			if err != nil {
				return err
			}
		} else {
			err = vmc.SetWithBsonByID(vm.ID, bson.D{{"host", host}})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func vmSnapshotUpdater() error {
	allVms, err := vm_repo.New().List()
	if err != nil {
		return err
	}

	for _, vm := range allVms {
		snapshotMap := fetchSnapshotStatus(&vm)
		if snapshotMap == nil {
			continue
		}

		err = vm_repo.New().SetSubsystem(vm.ID, "cs.snapshotMap", snapshotMap)
		if err != nil {
			return err
		}
	}

	return nil
}

func deploymentStatusUpdater() error {
	allDeployments, err := deployment_repo.New().List()
	if err != nil {
		return err
	}

	for _, deployment := range allDeployments {
		code, message, err := fetchDeploymentStatus(&deployment)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("error fetching deployment status: %w", err))
			continue
		}
		err = deployment_repo.New().SetWithBsonByID(deployment.ID, bson.D{{"statusCode", code}, {"statusMessage", message}})
		if err != nil {
			return err
		}
	}

	return nil
}

// deploymentPingUpdater is a worker that pings deployments.
// It stores the result in the database.
func deploymentPingUpdater() error {
	deployments, err := deployment_repo.New().List()
	if err != nil {
		return err
	}

	for _, deployment := range deployments {
		pingDeployment(&deployment)
	}

	return nil
}
