package logger

import (
	"context"
	"encoding/json"
	"errors"
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
	sErrors "go-deploy/service/errors"
	"go-deploy/utils"
	"os"
	"time"
)

// PodLogger is a worker that logs deployments.
func PodLogger(ctx context.Context) error {
	for _, zone := range config.Config.EnabledZones() {
		log.Println("Setting up log stream for zone", zone.Name)

		dZone := zone
		cancelFuncs := make(map[string]context.CancelFunc)
		err := message_queue.New().Consume(ctx, LogQueueKey(zone.Name), OnPodEvent(ctx, &dZone, cancelFuncs))
		if err != nil {
			return err
		}
	}

	return nil
}

func OnPodEvent(ctx context.Context, zone *configModels.Zone, cancelFuncs map[string]context.CancelFunc) func(bytes []byte) error {
	return func(bytes []byte) error {
		kvc := key_value.New()

		name, err := os.Hostname()
		if err != nil {
			return err
		}

		var logEvent LogEvent
		err = json.Unmarshal(bytes, &logEvent)
		if err != nil {
			return err
		}

		switch logEvent.PodEvent {
		case k8s.PodEventAdded:
			// We initially use a 2x lifetime to ensure that the logger is not removed while it is being set up
			didSet, err := kvc.SetNX(OwnerLogKey(logEvent.PodName, zone.Name), name, LoggerLifetime*2)
			if err != nil {
				return err
			}

			if !didSet {
				log.Printf("Pod %s is already being logged", logEvent.PodName)
				return nil
			}

			onLog := func(deploymentName string, lines []model.Log) {

				log.Printf("Adding %d logs for deployment %s", len(lines), deploymentName)

				err = deployment_repo.New().AddLogsByName(deploymentName, lines...)
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to add k8s logs for deployment %s. details: %w", logEvent.PodName, err))
					return
				}

				err = SetLastLogged(kvc, logEvent.PodName, zone.Name, time.Now())
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to set last logged time for pod %s. details: %w", logEvent.PodName, err))
					return
				}
			}

			log.Println("Setting up log stream for pod", logEvent.PodName)
			lastLogged := LastLogged(kvc, logEvent.PodName, zone.Name)
			loggerCtx, cancelFunc := context.WithCancel(context.Background())
			err = service.V2().Deployments().K8s().SetupPodLogStream(loggerCtx, zone, logEvent.PodName, lastLogged, onLog)
			if err != nil {
				cancelFunc()
				if errors.Is(err, sErrors.DeploymentNotFoundErr) {
					return nil
				}

				return err
			}

			cancelFuncs[logEvent.PodName] = cancelFunc

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
						didSet, err := kvc.SetXX(OwnerLogKey(logEvent.PodName, zone.Name), name, LoggerLifetime)
						if err != nil {
							utils.PrettyPrintError(fmt.Errorf("failed to update ownership of pod %s. details: %w", logEvent.PodName, err))
							return
						}

						if !didSet {
							log.Printf("Logger no longer owns pod %s. Cancelling", logEvent.PodName)
							if cancelFunc, ok := cancelFuncs[logEvent.PodName]; ok {
								cancelFunc()
								delete(cancelFuncs, logEvent.PodName)
							}
							return
						}
					}
					time.Sleep(1 * time.Second)
				}
			}(ctx, loggerCtx)
		case k8s.PodEventDeleted:
			log.Println("Removing log stream for pod", logEvent.PodName)
			if cancelFunc, ok := cancelFuncs[logEvent.PodName]; ok {
				cancelFunc()
				delete(cancelFuncs, logEvent.PodName)
			}
		}

		return nil
	}
}

func LastLogged(kvc *key_value.Client, podName, zoneName string) time.Time {
	// This might cause overlap, but it is better than missing logs
	fallback := time.Now().Add(-LoggerLifetime)

	val, err := kvc.Get(LastLogKey(podName, zoneName))
	if err != nil {
		return fallback
	}

	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return fallback
	}

	return t
}

func SetLastLogged(kvc *key_value.Client, podName, zoneName string, t time.Time) error {
	log.Printf("Setting last logged time for pod %s in zone %s to %s", podName, zoneName, t.Format(time.RFC3339))

	// Keep the entry for a week, so it clears up after a while
	return kvc.Set(LastLogKey(podName, zoneName), t.Format(time.RFC3339), time.Hour*24*7)
}
