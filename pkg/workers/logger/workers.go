package logger

import (
	"context"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/service/deployment_service/gitlab_service"
	"go-deploy/service/deployment_service/k8s_service"
	"go-deploy/utils"
	"log"
	"sync"
	"time"
)

func deploymentLogger(ctx context.Context) {
	defer log.Println("deploymentLogger stopped")

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
		case <-time.After(1000 * time.Millisecond):
			currentIDs := make([]string, len(current))
			idx := 0
			for id := range current {
				currentIDs[idx] = id
				idx++
			}

			ids, err := deploymentModel.New().ExcludeIDs(currentIDs...).ListIDs()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to list deployment ids. details: %w", err))
				continue
			}

			for _, idInList := range ids {
				id := idInList
				// setup log stream
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
						err = deploymentModel.New().AddLogs(id.ID, deploymentModel.Log{
							Source:    deploymentModel.LogSourcePod,
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
						utils.PrettyPrintError(fmt.Errorf("failed to setup deployment log stream for deployment %s. details: %w", id.ID, err))
						shouldCancel <- id.ID
						return
					}

					err = gitlab_service.SetupLogStream(logCtx, id.ID, func(line string, createdAt time.Time) {
						err = deploymentModel.New().AddLogs(id.ID, deploymentModel.Log{
							Source:    deploymentModel.LogSourceBuild,
							Prefix:    "[build]",
							Line:      line,
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
