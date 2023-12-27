package status_update

import (
	"fmt"
	deploymentModels "go-deploy/models/sys/deployment"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/cs"
	csModels "go-deploy/pkg/subsystems/cs/models"
)

func withClient(zoneName string) (*cs.Client, error) {
	zone := config.Config.VM.GetZone(zoneName)
	if zone == nil {
		return nil, fmt.Errorf("zone with id %s not found", zoneName)
	}

	return cs.New(&cs.ClientConf{
		URL:         config.Config.CS.URL,
		ApiKey:      config.Config.CS.ApiKey,
		Secret:      config.Config.CS.Secret,
		ZoneID:      zone.ZoneID,
		ProjectID:   zone.ProjectID,
		IpAddressID: zone.IpAddressID,
		NetworkID:   zone.NetworkID,
	})
}

func fetchCsStatus(vm *vmModels.VM) (int, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get status for cs vm %s. details: %w", vm.Name, err)
	}

	unknownMsg := status_codes.GetMsg(status_codes.ResourceUnknown)

	client, err := withClient(vm.Zone)
	if err != nil {
		return status_codes.ResourceUnknown, unknownMsg, makeError(err)
	}

	csVmID := vm.Subsystems.CS.VM.ID
	if csVmID == "" {
		return status_codes.ResourceNotFound, status_codes.GetMsg(status_codes.ResourceNotFound), nil
	}

	status, err := client.GetVmStatus(csVmID)
	if err != nil {
		return status_codes.ResourceUnknown, unknownMsg, makeError(err)
	}

	if status == "" {
		return status_codes.ResourceNotFound, status_codes.GetMsg(status_codes.ResourceNotFound), nil
	}

	var statusCode int
	switch status {
	case "Starting":
		statusCode = status_codes.ResourceStarting
	case "Running":
		statusCode = status_codes.ResourceRunning
	case "Stopping":
		statusCode = status_codes.ResourceStopping
	case "Stopped":
		statusCode = status_codes.ResourceStopped
	case "Migrating":
		statusCode = status_codes.ResourceRunning
	case "Error":
		statusCode = status_codes.ResourceError
	case "Unknown":
		statusCode = status_codes.ResourceUnknown
	case "Shutdown":
		statusCode = status_codes.ResourceStopped
	default:
		statusCode = status_codes.ResourceUnknown
	}

	return statusCode, status_codes.GetMsg(statusCode), nil
}

func fetchVmStatus(vm *vmModels.VM) (int, string, error) {
	if vm.DoingActivity(vmModels.ActivityCreatingSnapshot) {
		return status_codes.ResourceCreatingSnapshot, status_codes.GetMsg(status_codes.ResourceCreatingSnapshot), nil
	}

	if vm.DoingActivity(vmModels.ActivityApplyingSnapshot) {
		return status_codes.ResourceApplyingSnapshot, status_codes.GetMsg(status_codes.ResourceApplyingSnapshot), nil
	}

	csStatusCode, csStatusMessage, err := fetchCsStatus(vm)

	if csStatusCode == status_codes.ResourceUnknown || csStatusCode == status_codes.ResourceNotFound {
		if vm.BeingDeleted() {
			return status_codes.ResourceBeingDeleted, status_codes.GetMsg(status_codes.ResourceBeingDeleted), nil
		}

		if vm.BeingCreated() {
			return status_codes.ResourceBeingCreated, status_codes.GetMsg(status_codes.ResourceBeingCreated), nil
		}
	}

	if csStatusCode == status_codes.ResourceRunning && vm.BeingCreated() {
		return status_codes.ResourceBeingCreated, status_codes.GetMsg(status_codes.ResourceBeingCreated), nil
	}

	if csStatusCode == status_codes.ResourceRunning && vm.BeingDeleted() {
		return status_codes.ResourceStopping, status_codes.GetMsg(status_codes.ResourceStopping), nil
	}

	return csStatusCode, csStatusMessage, err
}

func fetchSnapshotStatus(vm *vmModels.VM) map[string]csModels.SnapshotPublic {
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

func fetchDeploymentStatus(deployment *deploymentModels.Deployment) (int, string, error) {

	if deployment == nil {
		return status_codes.ResourceNotFound, status_codes.GetMsg(status_codes.ResourceNotFound), nil
	}

	if deployment.BeingDeleted() {
		return status_codes.ResourceBeingDeleted, status_codes.GetMsg(status_codes.ResourceBeingDeleted), nil
	}

	if deployment.BeingCreated() {
		return status_codes.ResourceBeingCreated, status_codes.GetMsg(status_codes.ResourceBeingCreated), nil
	}

	if deployment.DoingActivity(deploymentModels.ActivityRestarting) {
		return status_codes.ResourceRestarting, status_codes.GetMsg(status_codes.ResourceRestarting), nil
	}

	if deployment.DoingActivity(deploymentModels.ActivityBuilding) {
		return status_codes.ResourceBuilding, status_codes.GetMsg(status_codes.ResourceBuilding), nil
	}

	return status_codes.ResourceRunning, status_codes.GetMsg(status_codes.ResourceRunning), nil
}
