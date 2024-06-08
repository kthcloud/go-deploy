package logger

import (
	"context"
	"fmt"
	configModels "go-deploy/models/config"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/key_value"
	"go-deploy/pkg/db/message_queue"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/service"
	"go-deploy/utils"
	"time"
)

type LogEvent struct {
	PodName  string
	PodEvent string
}

func PodLoggerControl(ctx context.Context) error {
	for _, zone := range config.Config.EnabledZones() {
		if !zone.HasCapability(configModels.ZoneCapabilityDeployment) {
			continue
		}

		log.Println("Setting up pod event listener for zone", zone.Name)

		mqc := message_queue.New()
		kvc := key_value.New()

		z := zone

		// Set up a listener for expired key events for every key that matches "logs:[a-z0-9-]"
		// This is used to ensure that a new logger is created for a pod if the previous one fails
		pattern := fmt.Sprintf("^logs:%s:[a-zA-Z0-9-]+$", zone.Name)
		println("Setting up expiration listener for pattern", pattern)
		err := kvc.SetUpExpirationListener(ctx, fmt.Sprintf("^logs:%s:[a-zA-Z0-9-]+$", zone.Name), func(key string) error {

			podName, zoneName := PodAndZoneNameFromLogKey(key)
			if zoneName != z.Name {
				// Ignore keys that are not for this zone
				return nil
			}

			// Check if Pod still exists
			exists, err := service.V2().Deployments().K8s().PodExists(&z, podName)
			if err != nil {
				return fmt.Errorf("failed to check if pod %s exists. details: %w", podName, err)
			}

			if !exists {
				// Clean up the keys
				log.Printf("Pod %s no longer exists in zone %s. Cleaning up keys", podName, z.Name)
				_ = kvc.Del(LogKey(podName, z.Name))
				_ = kvc.Del(LastLogKey(podName, z.Name))
				_ = kvc.Del(OwnerLogKey(podName, z.Name))
				_ = mqc.Publish(LogQueueKey(z.Name), LogEvent{
					PodName:  podName,
					PodEvent: k8s.PodEventDeleted,
				})
				return nil
			}

			// Reset the expired key so that it can be used again
			_, err = kvc.SetNX(key, false, LoggerLifetime)
			if err != nil {
				return fmt.Errorf("failed to reset expired key for pod %s. details: %w", podName, err)
			}

			count, err := mqc.GetListeners(LogQueueKey(zone.Name))
			if err != nil {
				return err
			}

			if count == 0 {
				log.Printf("No logger listeners active for zone %s. Retrying in %s (Pod: %s)", zone.Name, LoggerLifetime.String(), podName)
				return nil
			}

			// If a non-expired owner key exists, then the logger is still active
			isSet, err := kvc.IsSet(OwnerLogKey(podName, z.Name))
			if err != nil {
				return fmt.Errorf("failed to check if owner key is set for pod %s. details: %w", podName, err)
			}

			if isSet {
				return nil
			}

			// Publish the pod name to the active listeners
			err = mqc.Publish(LogQueueKey(z.Name), LogEvent{
				PodName:  podName,
				PodEvent: k8s.PodEventAdded,
			})
			if err != nil {
				return fmt.Errorf("failed to publish pod event for pod %s. details: %w", podName, err)
			}

			return nil
		})

		// Listen to pod events to set up loggers
		err = service.V2().Deployments().K8s().SetupPodWatcher(ctx, &z, func(podName string, event string) {
			switch event {
			case k8s.PodEventAdded:
			case k8s.PodEventUpdated:
			case k8s.PodEventDeleted:
				// We assume that the logger stops on its own
				// Clean up the keys
				_ = kvc.Del(LogKey(podName, z.Name))
				_ = kvc.Del(LastLogKey(podName, z.Name))
				_ = kvc.Del(OwnerLogKey(podName, z.Name))
				_ = mqc.Publish(LogQueueKey(z.Name), LogEvent{
					PodName:  podName,
					PodEvent: k8s.PodEventDeleted,
				})

				return
			}

			_, err = kvc.SetNX(LogKey(podName, z.Name), false, 1*time.Second)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to activate pod processing (when no listeners). details: %w", err))
				return
			}

			return
		})
		if err != nil {
			return fmt.Errorf("failed to set up deployment status watcher for zone %s. details: %w", zone.Name, err)
		}
	}

	return nil
}
