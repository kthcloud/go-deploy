package status_update

import (
	"context"
	"fmt"
	configModels "go-deploy/models/config"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/deployment_repo"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service/v1/deployments/k8s_service"
	"go.mongodb.org/mongo-driver/bson"
)

func deploymentStatusUpdater(ctx context.Context) error {
	for _, zone := range config.Config.Zones {
		if !zone.HasCapability(configModels.ZoneCapabilityDeployment) {
			continue
		}

		log.Println("Setting up deployment status watcher for zone", zone.Name)

		z := zone
		err := k8s_service.New().SetupStatusWatcher(ctx, &z, "deployment", func(name string, incomingStatus interface{}) {
			if status, ok := incomingStatus.(*models.DeploymentStatus); ok {
				deploymentStatus := parseDeploymentStatus(status)

				err := deployment_repo.New().SetWithBsonByName(name, bson.D{{"status", deploymentStatus}})
				if err != nil {
					log.Println("Failed to update VM status for", name, "with status", deploymentStatus, "details:", err)
					return
				}
			}
		})

		if err != nil {
			return fmt.Errorf("failed to setup vm status watcher for zone %s. details: %w", zone.Name, err)
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

func parseDeploymentStatus(status *models.DeploymentStatus) string {
	var statusCode int
	if status.ReadyReplicas == status.DesiredReplicas {
		statusCode = status_codes.ResourceRunning
	} else if status.Generation <= 1 && (status.ReadyReplicas != status.DesiredReplicas || status.UnavailableReplicas > 0) {
		statusCode = status_codes.ResourceCreating
	} else if status.DesiredReplicas == 0 && status.ReadyReplicas > 0 {
		statusCode = status_codes.ResourceStopping
	} else if status.DesiredReplicas == 0 {
		statusCode = status_codes.ResourceStopped
	} else if status.ReadyReplicas != status.DesiredReplicas {
		statusCode = status_codes.ResourceScaling
	}

	return status_codes.GetMsg(statusCode)
}
