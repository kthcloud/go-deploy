package status_update

import (
	"context"
	"fmt"
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/deployment_repo"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service/v1/deployments/k8s_service"
)

func DeploymentStatusListener(ctx context.Context) error {
	for _, zone := range config.Config.EnabledZones() {
		if !zone.HasCapability(configModels.ZoneCapabilityDeployment) {
			continue
		}

		log.Println("Setting up deployment status watcher for zone", zone.Name)

		z := zone
		err := k8s_service.New().SetupStatusWatcher(ctx, &z, "deployment", func(name string, incomingStatus interface{}) {
			if status, ok := incomingStatus.(*model.DeploymentStatus); ok {
				deploymentStatus := parseDeploymentStatus(status)

				err := deployment_repo.New().SetStatusByName(name, deploymentStatus, &model.ReplicaStatus{
					DesiredReplicas:     status.DesiredReplicas,
					ReadyReplicas:       status.ReadyReplicas,
					AvailableReplicas:   status.AvailableReplicas,
					UnavailableReplicas: status.UnavailableReplicas,
				})
				if err != nil {
					log.Println("Failed to set status for deployment", name, "details:", err)
					return
				}

				// If the status is running, we unset the errorStatus, since it no longer applies
				if deploymentStatus == status_codes.GetMsg(status_codes.ResourceRunning) {
					err = deployment_repo.New().UnsetErrorByName(name)
					if err != nil {
						log.Println("Failed to unset error status for", name, "details:", err)
					}
				}
			}
		})
		if err != nil {
			return fmt.Errorf("failed to set up deployment status watcher for zone %s. details: %w", zone.Name, err)
		}
	}

	return nil
}

func DeploymentEventListener(ctx context.Context) error {
	for _, zone := range config.Config.EnabledZones() {
		if !zone.HasCapability(configModels.ZoneCapabilityDeployment) {
			continue
		}

		log.Println("Setting up deployment event listener for zone", zone.Name)

		z := zone
		err := k8s_service.New().SetupStatusWatcher(ctx, &z, "event", func(name string, incomingStatus interface{}) {
			if event, ok := incomingStatus.(*model.DeploymentEvent); ok {
				if event.Type != models.EventTypeWarning {
					return
				}

				if event.ObjectKind != models.EventObjectKindDeployment {
					return
				}

				deploymentEventStatus := parseDeploymentErrorStatus(event)
				if deploymentEventStatus == status_codes.GetMsg(status_codes.ResourceUnknown) {
					return
				}

				err := deployment_repo.New().SetErrorByName(name, &model.DeploymentError{
					Reason:      deploymentEventStatus,
					Description: event.Description,
				})
				if err != nil {
					log.Println("Failed to set error status for deployment", name, "details:", err)
					return
				}
			}
		})
		if err != nil {
			return fmt.Errorf("failed to set up deployment event status watcher for zone %s. details: %w", zone.Name, err)
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

func parseDeploymentStatus(status *model.DeploymentStatus) string {
	var statusCode int
	if status.DesiredReplicas == 0 && status.ReadyReplicas == 0 && status.AvailableReplicas == 0 && status.UnavailableReplicas == 0 {
		statusCode = status_codes.ResourceDisabled
	} else if status.ReadyReplicas == status.DesiredReplicas {
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

func parseDeploymentErrorStatus(status *model.DeploymentEvent) string {
	switch status.Reason {
	case models.EventReasonMountFailed:
		return status_codes.GetMsg(status_codes.ResourceMountFailed)
	case models.EventReasonCrashLoop:
		return status_codes.GetMsg(status_codes.ResourceCrashLoop)
	case models.EventReasonImagePullFailed:
		return status_codes.GetMsg(status_codes.ResourceImagePullFailed)
	default:
		return status_codes.GetMsg(status_codes.ResourceUnknown)
	}
}
