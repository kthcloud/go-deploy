package logger

import (
	"context"
	"fmt"
	"time"

	configModels "github.com/kthcloud/go-deploy/models/config"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/db/key_value"
	"github.com/kthcloud/go-deploy/pkg/db/message_queue"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s"
	"github.com/kthcloud/go-deploy/service"
	"github.com/kthcloud/go-deploy/utils"
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

		})
		if err != nil {
			return fmt.Errorf("failed to set up deployment status watcher for zone %s. details: %w", zone.Name, err)
		}

		// Synchronize the existing pods with the loggers at an interval
		go func(ctx context.Context) {
			ticker := time.NewTicker(LoggerSynchronize)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					pods, err := service.V2().Deployments().K8s().Pods(&z)
					if err != nil {
						utils.PrettyPrintError(fmt.Errorf("failed to get pods for zone %s. details: %w", z.Name, err))
						continue
					}

					activePods, err := ActivePods(kvc, z.Name)
					if err != nil {
						utils.PrettyPrintError(fmt.Errorf("failed to get active pods for zone %s. details: %w", z.Name, err))
						continue
					}

					for _, pod := range pods {
						// If exists in active pods, skip
						if _, ok := activePods[pod.Name]; ok {
							delete(activePods, pod.Name)
							continue
						}

						// If not, mark it as added
						_, err := kvc.SetNX(LogKey(pod.Name, z.Name), false, LoggerLifetime)
						if err != nil {
							utils.PrettyPrintError(fmt.Errorf("failed to set log key for pod %s. details: %w", pod.Name, err))
							continue
						}
						delete(activePods, pod.Name)
					}

					// If there are any active pods left, mark them as deleted
					for podName := range activePods {
						_ = kvc.Del(LogKey(podName, z.Name))
						_ = kvc.Del(LastLogKey(podName, z.Name))
						_ = kvc.Del(OwnerLogKey(podName, z.Name))
						_ = mqc.Publish(LogQueueKey(z.Name), LogEvent{
							PodName:  podName,
							PodEvent: k8s.PodEventDeleted,
						})
					}
				}
			}
		}(ctx)
	}

	return nil
}
