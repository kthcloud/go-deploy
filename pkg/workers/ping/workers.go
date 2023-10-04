package ping

import (
	"context"
	"fmt"
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/pkg/conf"
	"go-deploy/service/deployment_service"
	"go-deploy/utils"
	"log"
	"net/http"
	"strings"
	"time"
)

func deploymentPingUpdater(ctx context.Context) {
	defer log.Println("deployment ping updater stopped")

	for {
		select {
		case <-time.After(time.Duration(conf.Env.Deployment.PingInterval) * time.Second):
			makeError := func(err error) error {
				return fmt.Errorf("error fetching deployments. details: %w", err)
			}

			deployments, err := deployment_service.GetAll()
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

func pingDeployment(deployment *deploymentModels.Deployment) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to ping deployments. details: %w", err)
	}

	baseURL := deployment.GetURL()
	if baseURL == nil {
		utils.PrettyPrintError(makeError(fmt.Errorf("deployment %s has no url", deployment.Name)))
		return
	}

	mainApp := deployment.GetMainApp()
	if mainApp == nil {
		utils.PrettyPrintError(makeError(fmt.Errorf("deployment %s has no main app", deployment.Name)))
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

func pingAndSave(deployment deploymentModels.Deployment, url string) {
	code, err := ping(url)

	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("error fetching deployment status ping. details: %w", err))
		return
	}

	_ = deployment_service.SavePing(deployment.ID, code)
}

func ping(url string) (int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("User-Agent", "kthcloud")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

func goodURL(url string) bool {
	rfc3986Characters := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~:/?#[]@!$&'()*+,;="
	for _, c := range url {
		if !strings.ContainsRune(rfc3986Characters, c) {
			return false
		}
	}
	return true
}
