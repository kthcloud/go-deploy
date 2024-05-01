package status_update

import (
	"context"
	"fmt"
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/vm_repo"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service/v2/vms/k8s_service"
	"go.mongodb.org/mongo-driver/bson"
)

func VmStatusListener(ctx context.Context) error {
	for _, zone := range config.Config.Zones {
		if !zone.HasCapability(configModels.ZoneCapabilityVM) {
			continue
		}

		log.Println("Setting up VM status watcher for zone", zone.Name)

		z := zone
		err := k8s_service.New().SetupStatusWatcher(ctx, &z, "vm", func(name string, incomingStatus interface{}) {
			if status, ok := incomingStatus.(*models.VmStatus); ok {
				kubeVirtStatus := parseVmStatus(status)
				err := vm_repo.New(version.V2).SetWithBsonByName(name, bson.D{{"status", kubeVirtStatus}})
				if err != nil {
					log.Printf("Failed to update VM status for %s. details: %s", name, err.Error())
					return
				}

				if kubeVirtStatus == status_codes.GetMsg(status_codes.ResourceStopped) {
					err = vm_repo.New(version.V2).UnsetByName(name, "host")
					if err != nil {
						log.Printf("Failed to update VM instance status for %s. details: %s", name, err.Error())
						return
					}
				}
			}
		})

		err = k8s_service.New().SetupStatusWatcher(ctx, &z, "vmi", func(name string, incomingStatus interface{}) {
			if status, ok := incomingStatus.(*models.VmiStatus); ok {
				if status.Host == nil {
					err = vm_repo.New(version.V2).UnsetByName(name, "host")
					if err != nil {
						log.Printf("Failed to update VM instance status for %s. details: %s", name, err.Error())
						return
					}
				} else {
					err = vm_repo.New(version.V2).SetWithBsonByName(name, bson.D{{"host", model.Host{Name: *status.Host}}})
					if err != nil {
						log.Printf("Failed to update VM instance status for %s. details: %s", name, err.Error())
						return
					}
				}

			}
		})

		if err != nil {
			return fmt.Errorf("failed to setup vm status watcher for zone %s. details: %w", zone.Name, err)
		}
	}

	return nil
}

func parseVmStatus(status *models.VmStatus) string {
	var statusCode int
	switch status.PrintableStatus {
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
		statusCode = status_codes.ResourceDeleting
	case "CrashLoopBackOff", "Unknown", "Unschedulable", "ErrImagePull", "ImagePullBackOff", "PvcNotFound", "DataVolumeError":
		statusCode = status_codes.ResourceError
	default:
		statusCode = status_codes.ResourceUnknown
	}

	return status_codes.GetMsg(statusCode)
}
