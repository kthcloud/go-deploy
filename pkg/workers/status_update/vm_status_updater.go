package status_update

import (
	"context"
	"fmt"
	configModels "go-deploy/models/config"
	"go-deploy/models/version"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/vm_repo"
	"go-deploy/pkg/log"
	"go-deploy/service/v2/vms/k8s_service"
	"go.mongodb.org/mongo-driver/bson"
)

func vmStatusUpdaterV2(ctx context.Context) error {
	for _, zone := range config.Config.Deployment.Zones {
		if !zone.HasCapability(configModels.CapabilityKubeVirt) {
			continue
		}

		log.Println("Setting up VM status watcher for zone", zone.Name)

		z := zone
		err := k8s_service.New().SetupStatusWatcher(ctx, &z, "vm", func(name, kubeVirtStatus string) {
			status := parseKubeVirtStatus(kubeVirtStatus)

			err := vm_repo.New(version.V2).SetWithBsonByName(name, bson.D{{"status", status}})
			if err != nil {
				log.Println("Failed to update VM status for", name, "with status", status, "details:", err)
				return
			}
		})

		if err != nil {
			return fmt.Errorf("failed to setup vm status watcher for zone %s. details: %w", zone.Name, err)
		}
	}

	return nil
}

func parseKubeVirtStatus(status string) string {
	var statusCode int
	switch status {
	case "Provisioning", "WaitingForVolumeBinding":
		statusCode = status_codes.ResourceProvisioning
	case "Starting":
		statusCode = status_codes.ResourceStarting
	case "Running":
		statusCode = status_codes.ResourceRunning
	case "Migrating":
		statusCode = status_codes.ResourceMigrating
	case "Stopped", "Paused":
		statusCode = status_codes.ResourceStopped
	case "Stopping":
		statusCode = status_codes.ResourceStopping
	case "Terminating":
		statusCode = status_codes.ResourceBeingDeleted
	case "CrashLoopBackOff", "Unknown", "Unschedulable", "ErrImagePull", "ImagePullBackOff", "PvcNotFound", "DataVolumeError":
		statusCode = status_codes.ResourceError
	default:
		statusCode = status_codes.ResourceUnknown
	}

	return status_codes.GetMsg(statusCode)
}
