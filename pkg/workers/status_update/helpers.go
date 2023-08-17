package status_update

import (
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/subsystems/cs"
	csModels "go-deploy/pkg/subsystems/cs/models"
)

func withClient() (*cs.Client, error) {
	return cs.New(&cs.ClientConf{
		URL:         conf.Env.CS.URL,
		ApiKey:      conf.Env.CS.ApiKey,
		Secret:      conf.Env.CS.Secret,
		IpAddressID: conf.Env.CS.IpAddressID,
		NetworkID:   conf.Env.CS.NetworkID,
		ProjectID:   conf.Env.CS.ProjectID,
		ZoneID:      conf.Env.CS.ZoneID,
	})
}

func fetchCsStatus(vm *vmModel.VM) (int, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get status for cs vm %s. details: %s", vm.Name, err)
	}

	unknownMsg := status_codes.GetMsg(status_codes.ResourceUnknown)

	client, err := withClient()
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

func fetchVmStatus(vm *vmModel.VM) (int, string, error) {
	if vm.DoingActivity(vmModel.ActivityCreatingSnapshot) {
		return status_codes.ResourceCreatingSnapshot, status_codes.GetMsg(status_codes.ResourceCreatingSnapshot), nil
	}

	if vm.DoingActivity(vmModel.ActivityApplyingSnapshot) {
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

func fetchSnapshotStatus(vm *vmModel.VM) map[string]csModels.SnapshotPublic {
	client, err := withClient()
	if err != nil {
		return nil
	}

	snapshots, err := client.ReadAllSnapshots(vm.Subsystems.CS.VM.ID)
	if err != nil {
		return nil
	}

	snapshotMap := make(map[string]csModels.SnapshotPublic)
	for _, snapshot := range snapshots {
		snapshotMap[snapshot.Name] = snapshot
	}

	return snapshotMap
}

func fetchDeploymentStatus(deployment *deploymentModel.Deployment) (int, string, error) {

	if deployment == nil {
		return status_codes.ResourceNotFound, status_codes.GetMsg(status_codes.ResourceNotFound), nil
	}

	if deployment.BeingDeleted() {
		return status_codes.ResourceBeingDeleted, status_codes.GetMsg(status_codes.ResourceBeingDeleted), nil
	}

	if deployment.BeingCreated() {
		return status_codes.ResourceBeingCreated, status_codes.GetMsg(status_codes.ResourceBeingCreated), nil
	}

	if deployment.DoingActivity(deploymentModel.ActivityRestarting) {
		return status_codes.ResourceRestarting, status_codes.GetMsg(status_codes.ResourceRestarting), nil
	}

	if deployment.DoingActivity(deploymentModel.ActivityBuilding) {
		return status_codes.ResourceBuilding, status_codes.GetMsg(status_codes.ResourceBuilding), nil
	}

	return status_codes.ResourceRunning, status_codes.GetMsg(status_codes.ResourceRunning), nil
}
