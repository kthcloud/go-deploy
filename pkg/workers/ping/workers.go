package ping

import (
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/pkg/sys"
	"go-deploy/service/deployment_service"
	"log"
	"net/http"
	"time"
)

func deploymentPingUpdater(ctx *sys.Context) {
	log.Println("starting deployment ping updater")
	for {
		if ctx.Stop {
			break
		}

		updateAllDeploymentPings()
		time.Sleep(30 * time.Second)
	}
}

func updateAllDeploymentPings() {
	deployments, _ := deployment_service.GetAll()

	for _, deployment := range deployments {

		url := getURL(&deployment)

		if url == nil {
			log.Println("deployment has no url")
			continue
		}

		go updateOneDeploymentPing(&deployment, *url)
	}
}

func updateOneDeploymentPing(deployment *deploymentModels.Deployment, url string) {
	code, err := ping(url)

	if err != nil {
		log.Println("error fetching deployment status ping. details:", err)
	}

	_ = deployment_service.SavePing(deployment.ID, code)
}

func ping(url string) (int, error) {
	resp, err := http.Get("https://" + url)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

func getURL(deployment *deploymentModels.Deployment) *string {
	if len(deployment.Subsystems.K8s.Ingress.Hosts) > 0 && len(deployment.Subsystems.K8s.Ingress.Hosts[0]) > 0 {
		return &deployment.Subsystems.K8s.Ingress.Hosts[0]
	}
	return nil
}
