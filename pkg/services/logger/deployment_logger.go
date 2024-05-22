package logger

import (
	"context"
	"encoding/json"
	"fmt"
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/key_value"
	"go-deploy/pkg/db/message_queue"
	"go-deploy/pkg/db/resources/deployment_repo"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/service"
	"go-deploy/utils"
	"time"
)

// DeploymentLogger is a worker that logs deployments.
func DeploymentLogger(ctx context.Context) error {
	for _, zone := range config.Config.EnabledZones() {
		log.Println("Setting up log stream for zone", zone.Name)

		dZone := zone
		cancelFuncs := make(map[string]context.CancelFunc)
		err := message_queue.New().Consume(LogQueueKey(zone.Name), OnPodEvent(ctx, &dZone, cancelFuncs))
		if err != nil {
			return err
		}
	}

	return nil
}

func OnPodEvent(ctx context.Context, zone *configModels.Zone, cancelFuncs map[string]context.CancelFunc) func(bytes []byte) error {
	return func(bytes []byte) error {
		kvc := key_value.New()

		var logEvent LogEvent
		err := json.Unmarshal(bytes, &logEvent)
		if err != nil {
			return err
		}

		switch logEvent.PodEvent {
		case k8s.PodEventAdded:
			// We initially use a 2x lifetime to ensure that the logger is not removed while it is being set up
			didSet, err := kvc.SetNX(logEvent.PodName, true, LoggerLifetime*2)
			if err != nil {
				return err
			}

			if !didSet {
				log.Printf("Pod %s is already being logged", logEvent.PodName)
				return nil
			}

			log.Println("Setting up log stream for pod", logEvent.PodName)

			lastLogged := LastLogged(kvc, logEvent.PodName)
			onLog := func(deploymentName string, lines []model.Log) {
				err = deployment_repo.New().AddLogsByName(deploymentName, lines...)
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to add k8s logs for deployment %s. details: %w", logEvent.PodName, err))
					return
				}

				err = SetLastLogged(kvc, logEvent.PodName, time.Now())
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to set last logged time for pod %s. details: %w", logEvent.PodName, err))
					return
				}
			}

			loggerCtx, cancelFunc := context.WithCancel(context.Background())
			cancelFuncs[logEvent.PodName] = cancelFunc

			err = service.V1().Deployments().K8s().SetupPodLogStream(loggerCtx, zone, logEvent.PodName, lastLogged, onLog)
			if err != nil {
				return err
			}

			// Set up a loop to update ownership of pod name
			go func(ctx, loggerCtx context.Context) {
				tick := time.Tick(LoggerUpdate)
				for {
					select {
					case <-ctx.Done():
						return
					case <-loggerCtx.Done():
						return
					case <-tick:
						err = kvc.Set(logEvent.PodName, true, LoggerLifetime)
						if err != nil {
							utils.PrettyPrintError(fmt.Errorf("failed to update ownership of pod %s. details: %w", logEvent.PodName, err))
							return
						}
					}
				}
			}(ctx, loggerCtx)
		case k8s.PodEventDeleted:
			log.Println("Removing log stream for pod", logEvent.PodName)

			cancelFuncs[logEvent.PodName]()
			delete(cancelFuncs, logEvent.PodName)
		}

		return nil
	}
}

func LastLogged(kvc *key_value.Client, podName string) time.Time {
	// This might cause overlap, but it is better than missing logs
	fallback := time.Now().Add(-LoggerLifetime)

	val, err := kvc.Get(LastLogKey(podName))
	if err != nil {
		return fallback
	}

	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return fallback
	}

	return t
}

func SetLastLogged(kvc *key_value.Client, podName string, t time.Time) error {
	// Keep the entry for a week so it clears up after a while
	return kvc.Set(LastLogKey(podName), t.Format(time.RFC3339), time.Hour*24*7)
}
