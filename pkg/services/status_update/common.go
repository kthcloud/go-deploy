package status_update

import (
	"fmt"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/deployment_repo"
	"github.com/kthcloud/go-deploy/utils"
	"net/http"
	"strings"
)

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

	baseURL := deployment.GetURL(nil)
	if mainApp.Visibility != model.VisibilityPublic || mainApp.PingPath == "" || baseURL == nil {
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
		code = 0
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
