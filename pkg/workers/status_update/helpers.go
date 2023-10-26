package status_update

import (
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	vmModel "go-deploy/models/sys/vm"
	status_codes2 "go-deploy/pkg/app/status_codes"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/cs"
	csModels "go-deploy/pkg/subsystems/cs/models"
)

func withClient(zoneName string) (*cs.Client, error) {
	zone := conf.Env.VM.GetZone(zoneName)
	if zone == nil {
		return nil, fmt.Errorf("zone with id %s not found", zoneName)
	}

	return cs.New(&cs.ClientConf{
		URL:         conf.Env.CS.URL,
		ApiKey:      conf.Env.CS.ApiKey,
		Secret:      conf.Env.CS.Secret,
		ZoneID:      zone.ZoneID,
		ProjectID:   zone.ProjectID,
		IpAddressID: zone.IpAddressID,
		NetworkID:   zone.NetworkID,
	})
}

func fetchCsStatus(vm *vmModel.VM) (int, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get status for cs vm %s. details: %w", vm.Name, err)
	}

	unknownMsg := status_codes2.GetMsg(status_codes2.ResourceUnknown)

	client, err := withClient(vm.Zone)
	if err != nil {
		return status_codes2.ResourceUnknown, unknownMsg, makeError(err)
	}

	csVmID := vm.Subsystems.CS.VM.ID
	if csVmID == "" {
		return status_codes2.ResourceNotFound, status_codes2.GetMsg(status_codes2.ResourceNotFound), nil
	}

	status, err := client.GetVmStatus(csVmID)
	if err != nil {
		return status_codes2.ResourceUnknown, unknownMsg, makeError(err)
	}

	if status == "" {
		return status_codes2.ResourceNotFound, status_codes2.GetMsg(status_codes2.ResourceNotFound), nil
	}

	var statusCode int
	switch status {
	case "Starting":
		statusCode = status_codes2.ResourceStarting
	case "Running":
		statusCode = status_codes2.ResourceRunning
	case "Stopping":
		statusCode = status_codes2.ResourceStopping
	case "Stopped":
		statusCode = status_codes2.ResourceStopped
	case "Migrating":
		statusCode = status_codes2.ResourceRunning
	case "Error":
		statusCode = status_codes2.ResourceError
	case "Unknown":
		statusCode = status_codes2.ResourceUnknown
	case "Shutdown":
		statusCode = status_codes2.ResourceStopped
	default:
		statusCode = status_codes2.ResourceUnknown
	}

	return statusCode, status_codes2.GetMsg(statusCode), nil
}

func fetchVmStatus(vm *vmModel.VM) (int, string, error) {
	if vm.DoingActivity(vmModel.ActivityCreatingSnapshot) {
		return status_codes2.ResourceCreatingSnapshot, status_codes2.GetMsg(status_codes2.ResourceCreatingSnapshot), nil
	}

	if vm.DoingActivity(vmModel.ActivityApplyingSnapshot) {
		return status_codes2.ResourceApplyingSnapshot, status_codes2.GetMsg(status_codes2.ResourceApplyingSnapshot), nil
	}

	csStatusCode, csStatusMessage, err := fetchCsStatus(vm)

	if csStatusCode == status_codes2.ResourceUnknown || csStatusCode == status_codes2.ResourceNotFound {
		if vm.BeingDeleted() {
			return status_codes2.ResourceBeingDeleted, status_codes2.GetMsg(status_codes2.ResourceBeingDeleted), nil
		}

		if vm.BeingCreated() {
			return status_codes2.ResourceBeingCreated, status_codes2.GetMsg(status_codes2.ResourceBeingCreated), nil
		}
	}

	if csStatusCode == status_codes2.ResourceRunning && vm.BeingCreated() {
		return status_codes2.ResourceBeingCreated, status_codes2.GetMsg(status_codes2.ResourceBeingCreated), nil
	}

	if csStatusCode == status_codes2.ResourceRunning && vm.BeingDeleted() {
		return status_codes2.ResourceStopping, status_codes2.GetMsg(status_codes2.ResourceStopping), nil
	}

	return csStatusCode, csStatusMessage, err
}

func fetchSnapshotStatus(vm *vmModel.VM) map[string]csModels.SnapshotPublic {
	client, err := withClient(vm.Zone)
	if err != nil {
		return nil
	}

	snapshots, err := client.ReadAllSnapshots(vm.Subsystems.CS.VM.ID)
	if err != nil {
		return nil
	}

	snapshotMap := make(map[string]csModels.SnapshotPublic)
	for _, snapshot := range snapshots {
		snapshotMap[snapshot.ID] = snapshot
	}

	return snapshotMap
}

func fetchDeploymentStatus(deployment *deploymentModel.Deployment) (int, string, error) {

	if deployment == nil {
		return status_codes2.ResourceNotFound, status_codes2.GetMsg(status_codes2.ResourceNotFound), nil
	}

	if deployment.BeingDeleted() {
		return status_codes2.ResourceBeingDeleted, status_codes2.GetMsg(status_codes2.ResourceBeingDeleted), nil
	}

	if deployment.BeingCreated() {
		return status_codes2.ResourceBeingCreated, status_codes2.GetMsg(status_codes2.ResourceBeingCreated), nil
	}

	if deployment.DoingActivity(deploymentModel.ActivityRestarting) {
		return status_codes2.ResourceRestarting, status_codes2.GetMsg(status_codes2.ResourceRestarting), nil
	}

	if deployment.DoingActivity(deploymentModel.ActivityBuilding) {
		return status_codes2.ResourceBuilding, status_codes2.GetMsg(status_codes2.ResourceBuilding), nil
	}

	return status_codes2.ResourceRunning, status_codes2.GetMsg(status_codes2.ResourceRunning), nil
}
