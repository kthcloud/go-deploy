package ping

import (
	"context"
	"fmt"
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/pkg/config"
	"go-deploy/pkg/workers"
	"go-deploy/service/deployment_service"
	"go-deploy/service/deployment_service/client"
	"go-deploy/utils"
	"net/http"
	"strings"
	"time"
)

// deploymentPingUpdater is a worker that pings deployments.
// It stores the result in the database.
func deploymentPingUpdater(ctx context.Context) {
	defer workers.OnStop("deployment ping updater")

	reportTick := time.Tick(1 * time.Second)
	tick := time.Tick(time.Duration(config.Config.Deployment.PingInterval) * time.Second)

	for {
		select {
		case <-reportTick:
			workers.ReportUp("deploymentPingUpdater")

		case <-tick:
			makeError := func(err error) error {
				return fmt.Errorf("error fetching deployments. details: %w", err)
			}

			deployments, err := deployment_service.New().List(&client.ListOptions{})
			if err != nil {
				utils.PrettyPrintError(makeError(err))
				return
			}

			for _, deployment := range deployments {
				pingDeployment(&deployment)
			}
		case <-ctx.Done():
			return
		}
	}
}

// pingDeployment pings a deployment and stores the result in the database.
func pingDeployment(deployment *deploymentModels.Deployment) {
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
func pingAndSave(deployment deploymentModels.Deployment, url string) {
	code, err := ping(url)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("error fetching deployment status ping. details: %w", err))
		return
	}

	err = deployment_service.New().SavePing(deployment.ID, code)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("error saving deployment status ping. details: %w", err))
		return
	}
}

// resetPing resets the ping status of a deployment.
func resetPing(deployment deploymentModels.Deployment) {
	err := deployment_service.New().SavePing(deployment.ID, 0)
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
