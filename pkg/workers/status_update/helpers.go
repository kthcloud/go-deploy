package status_update

import (
	"fmt"
	"go-deploy/models/model"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/deployment_repo"
	"go-deploy/pkg/db/resources/job_repo"
	"go-deploy/pkg/subsystems/cs"
	csModels "go-deploy/pkg/subsystems/cs/models"
	"go-deploy/utils"
	"net/http"
	"strings"
)

// withClient is a helper function that return as CloudStack client.
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

// parseCsStatus is a helper function that parses the status of a VM from CloudStack.
func parseCsStatus(status string) (int, string) {
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

	return statusCode, status_codes.GetMsg(statusCode)
}

// fetchVmStatusV1 fetches the status of a VM.
func fetchVmStatusV1(vm *model.VM, csStatus string) (int, string, error) {
	csStatusCode, csStatusMessage := parseCsStatus(csStatus)

	// In case the "delete CS VM" part fails, the VM will be stuck in "Running" state.
	// So we need to check if the VM is being deleted and if so, return the correct status
	// to indicate that the VM is in fact being deleted.
	if csStatusMessage == "Running" && vm.BeingDeleted() {
		return status_codes.ResourceBeingDeleted, status_codes.GetMsg(status_codes.ResourceBeingDeleted), nil
	}

	anyCreateSnapshotJob, err := job_repo.New().
		FilterArgs("id", vm.ID).
		IncludeTypes(model.JobCreateVmUserSnapshot, model.JobCreateSystemVmSnapshot).
		IncludeStatus(model.JobStatusRunning).
		ExistsAny()
	if err != nil {
		return status_codes.ResourceUnknown, status_codes.GetMsg(status_codes.ResourceUnknown), fmt.Errorf("failed to check if snapshot job exists. details: %w", err)
	}

	if anyCreateSnapshotJob {
		return status_codes.ResourceCreatingSnapshot, status_codes.GetMsg(status_codes.ResourceCreatingSnapshot), nil
	}

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

// fetchVmStatusV2 fetches the status of a VM.
func fetchVmStatusV2(vm *model.VM) (int, string, error) {
	if vm.BeingCreated() {
		return status_codes.ResourceBeingCreated, status_codes.GetMsg(status_codes.ResourceBeingCreated), nil
	}

	if vm.BeingDeleted() {
		return status_codes.ResourceStopping, status_codes.GetMsg(status_codes.ResourceStopping), nil
	}

	return status_codes.ResourceRunning, status_codes.GetMsg(status_codes.ResourceRunning), nil
}

// fetchSnapshotStatus fetches the status of a VM's snapshots.
func fetchSnapshotStatus(vm *model.VM) map[string]csModels.SnapshotPublic {
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

// fetchDeploymentStatus fetches the status of a deployment.
func fetchDeploymentStatus(deployment *model.Deployment) (int, string, error) {

	if deployment == nil {
		return status_codes.ResourceNotFound, status_codes.GetMsg(status_codes.ResourceNotFound), nil
	}

	if deployment.BeingDeleted() {
		return status_codes.ResourceBeingDeleted, status_codes.GetMsg(status_codes.ResourceBeingDeleted), nil
	}

	if deployment.BeingCreated() {
		return status_codes.ResourceBeingCreated, status_codes.GetMsg(status_codes.ResourceBeingCreated), nil
	}

	if deployment.GetMainApp().Replicas == 0 {
		return status_codes.ResourceStopped, status_codes.GetMsg(status_codes.ResourceStopped), nil
	}

	if deployment.DoingActivity(model.ActivityRestarting) {
		return status_codes.ResourceRestarting, status_codes.GetMsg(status_codes.ResourceRestarting), nil
	}

	if deployment.DoingActivity(model.ActivityBuilding) {
		return status_codes.ResourceBuilding, status_codes.GetMsg(status_codes.ResourceBuilding), nil
	}

	return status_codes.ResourceRunning, status_codes.GetMsg(status_codes.ResourceRunning), nil
}

// pingDeployment pings a deployment and stores the result in the database.
func pingDeployment(deployment *model.Deployment) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to ping deployment. details: %w", err)
	}

	mainApp := deployment.GetMainApp()
	if mainApp == nil {
		utils.PrettyPrintError(makeError(fmt.Errorf("deployment %s has no main app", deployment.Name)))
		return
	}

	baseURL := deployment.GetURL()
	if mainApp.Private || mainApp.PingPath == "" || baseURL == nil {
		go resetPing(*deployment)
		return
	}

	subPath := mainApp.PingPath
	if len(mainApp.PingPath) > 0 && !strings.HasPrefix(mainApp.PingPath, "/") {
		subPath = "/" + subPath
	}

	pingURL := *baseURL + subPath
	if !goodURL(pingURL) {
		utils.PrettyPrintError(makeError(fmt.Errorf("deployment %s has invalid ping url %s", deployment.Name, pingURL)))
		return
	}

	go pingAndSave(*deployment, pingURL)
}

// pingAndSave pings a deployment and stores the result in the database.
func pingAndSave(deployment model.Deployment, url string) {
	code, err := ping(url)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("error fetching deployment status ping. details: %w", err))
		return
	}

	err = deployment_repo.New().SetPingResult(deployment.ID, code)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("error saving deployment status ping. details: %w", err))
		return
	}
}

// resetPing resets the ping status of a deployment.
func resetPing(deployment model.Deployment) {
	err := deployment_repo.New().SetPingResult(deployment.ID, 0)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("error resetting deployment status ping. details: %w", err))
		return
	}
}

// ping pings a url and returns the status code.
func ping(url string) (int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("User-Agent", "kthcloud")

	httpClient := &http.Client{}

	resp, err := httpClient.Do(req)
	if err != nil {
		// Check if certificate is invalid
		if strings.Contains(err.Error(), "x509: certificate") {
			return 400, nil
		}

		return 0, err
	}

	return resp.StatusCode, nil
}

// goodURL is a helper function that checks if a URL is valid.
func goodURL(url string) bool {
	rfc3986Characters := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~:/?#[]@!$&'()*+,;="
	for _, c := range url {
		if !strings.ContainsRune(rfc3986Characters, c) {
			return false
		}
	}
	return true
}
