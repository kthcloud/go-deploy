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

func PodEventListener(ctx context.Context) error {
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
		err := kvc.SetUpExpirationListener(ctx, "^logs:[a-zA-Z0-9-]+$", func(key string) error {
			podName := PodNameFromLogKey(key)

			// Check if Pod still exists
			exists, err := service.V2().Deployments().K8s().PodExists(&z, podName)
			if err != nil {
				return err
			}

			if !exists {
				// Clean up the keys
				_ = kvc.Del(LogKey(podName))
				_ = kvc.Del(LastLogKey(podName))
				_ = kvc.Del(OwnerLogKey(podName))
				_ = mqc.Publish(LogQueueKey(zone.Name), LogEvent{
					PodName:  podName,
					PodEvent: k8s.PodEventDeleted,
				})
				return nil
			}

			// Reset the expired key so that it can be used again
			_, err = kvc.SetNX(key, false, LoggerLifetime)
			if err == nil {
				return err
			}

			// Check if there are any active listeners, otherwise mark this pod as being processed
			count, err := mqc.GetListeners(LogQueueKey(zone.Name))
			if err != nil {
				return err
			}

			if count == 0 {
				log.Printf("No logger listeners active for zone %s. Retrying in %s (Pod: %s)", zone.Name, LoggerLifetime.String(), podName)
				return nil
			}

			// If n non-expired owner key exists, then the logger is still active
			isSet, err := kvc.IsSet(OwnerLogKey(podName))
			if err != nil {
				return err
			}

			if isSet {
				return nil
			}

			// Publish the pod name to the active listeners
			err = mqc.Publish(LogQueueKey(zone.Name), LogEvent{
				PodName:  podName,
				PodEvent: k8s.PodEventAdded,
			})
			if err != nil {
				return err
			}

			return nil
		})

		err = service.V2().Deployments().K8s().SetupPodWatcher(ctx, &z, func(podName string, event string) {
			switch event {
			case k8s.PodEventAdded:
			case k8s.PodEventUpdated:
			case k8s.PodEventDeleted:
				// We assume that the logger stops on its own
				// Clean up the keys
				_ = kvc.Del(LogKey(podName))
				_ = kvc.Del(LastLogKey(podName))
				_ = kvc.Del(OwnerLogKey(podName))
				_ = mqc.Publish(LogQueueKey(zone.Name), LogEvent{
					PodName:  podName,
					PodEvent: k8s.PodEventDeleted,
				})

				return
			}

			_, err = kvc.SetNX(LogKey(podName), false, 1*time.Second)
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
