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
	"time"
)

func deploymentPingUpdater(ctx context.Context) {
	defer log.Println("deployment ping updater stopped")

	for {
		select {
		case <-time.After(time.Duration(conf.Env.Deployment.PingInterval) * time.Second):
			updateAllDeploymentPings()
		case <-ctx.Done():
			return
		}
	}
}

func updateAllDeploymentPings() {
	deployments, _ := deployment_service.GetAll()

	log.Println("pinging", len(deployments), "deployments")

	for _, deployment := range deployments {

		url := getURL(&deployment)

		if url == nil {
			log.Println("deployment", deployment.Name, "has no url")
			continue
		}

		go updateOneDeploymentPing(deployment, *url)
	}
}

func updateOneDeploymentPing(deployment deploymentModels.Deployment, url string) {
	code, err := ping(url)

	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("error fetching deployment status ping. details: %w", err))
	}

	_ = deployment_service.SavePing(deployment.ID, code)
}

func ping(url string) (int, error) {
	req, err := http.NewRequest("GET", "https://"+url, nil)
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

func getURL(deployment *deploymentModels.Deployment) *string {
	ingress, ok := deployment.Subsystems.K8s.IngressMap["main"]
	if !ok || !ingress.Created() {
		return nil
	}

	if len(ingress.Hosts) > 0 && len(ingress.Hosts[0]) > 0 {
		return &ingress.Hosts[0]
	}
	return nil
}
