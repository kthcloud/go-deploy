package logger

import (
	"context"
	"errors"
	"fmt"
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/pkg/workers"
	"go-deploy/service/deployment_service/gitlab_service"
	"go-deploy/service/deployment_service/k8s_service"
	sErrors "go-deploy/service/errors"
	"go-deploy/utils"
	"log"
	"sync"
	"time"
)

func deploymentLogger(ctx context.Context) {
	defer workers.OnStop("deploymentLogger")

	mut := sync.Mutex{}

	current := make(map[string]bool)
	cancelFuncs := make(map[string]context.CancelFunc)
	shouldCancel := make(chan string)

	// listen for cancel requests
	go func() {
		for {
			select {
			case id := <-shouldCancel:
				go func() {
					mut.Lock()
					defer mut.Unlock()

					if cancel, ok := cancelFuncs[id]; ok {
						cancel()
						delete(cancelFuncs, id)
					}
					delete(current, id)
				}()
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-time.After(1 * time.Second):
			workers.ReportUp("deploymentLogger")

		case <-time.After(1000 * time.Millisecond):
			currentIDs := make([]string, len(current))
			idx := 0
			for id := range current {
				currentIDs[idx] = id
				idx++
			}

			ids, err := deploymentModels.New().
				ExcludeIDs(currentIDs...).
				WithoutActivities(deploymentModels.ActivityBeingCreated, deploymentModels.ActivityBeingDeleted).
				ListIDs()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to list deployment ids. details: %w", err))
				continue
			}

			for _, idInList := range ids {
				id := idInList

				// Setup log stream
				go func() {
					mut.Lock()
					defer mut.Unlock()

					if current[id.ID] {
						return
					}

					current[id.ID] = true
					logCtx, cancel := context.WithCancel(ctx)
					cancelFuncs[id.ID] = cancel

					err = k8s_service.New(nil).SetupLogStream(logCtx, id.ID, func(line string, podNumber int, createdAt time.Time) {
						err = deploymentModels.New().AddLogs(id.ID, deploymentModels.Log{
							Source:    deploymentModels.LogSourcePod,
							Prefix:    fmt.Sprintf("[pod %d]", podNumber),
							Line:      line,
							CreatedAt: createdAt,
						})
						if err != nil {
							utils.PrettyPrintError(fmt.Errorf("failed to add k8s logs for deployment %s. details: %w", id.ID, err))
							shouldCancel <- id.ID
							return
						}
					})

					if err != nil {
						if errors.Is(err, sErrors.BadStateErr) {
							log.Println("deployment", id.ID, "is not ready to setup log stream, will try again later")
							shouldCancel <- id.ID
							return
						}

						utils.PrettyPrintError(fmt.Errorf("failed to setup deployment log stream for deployment %s. details: %w", id.ID, err))
						shouldCancel <- id.ID
						return
					}

					err = gitlab_service.SetupLogStream(logCtx, id.ID, func(line string, createdAt time.Time) {
						err = deploymentModels.New().AddLogs(id.ID, deploymentModels.Log{
							Source: deploymentModels.LogSourceBuild,
							Prefix: "[build]",
							// Since this is sent as a string, and not a JSON object, we need to prepend the createdAt
							Line:      fmt.Sprintf("%s %s", createdAt.Format(time.RFC3339), line),
							CreatedAt: createdAt,
						})
						if err != nil {
							utils.PrettyPrintError(fmt.Errorf("failed to add gitlab logs for deployment %s. details: %w", id.ID, err))
							shouldCancel <- id.ID
							return
						}
					})

					if err != nil {
						utils.PrettyPrintError(fmt.Errorf("failed to setup gitlab log stream for deployment %s. details: %w", id.ID, err))
						shouldCancel <- id.ID
						return
					}
				}()
			}
		case <-ctx.Done():
			return
		}
	}
}
