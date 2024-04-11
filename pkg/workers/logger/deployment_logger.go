package logger

import (
	"context"
	"fmt"
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/deployment_repo"
	"go-deploy/pkg/log"
	"go-deploy/service/v1/deployments/k8s_service"
	"go-deploy/utils"
	"time"
)

// deploymentLogger is a worker that logs deployments.
func deploymentLogger(ctx context.Context) error {
	// Get allowed names
	allowedNames, err := deployment_repo.New().ListNames()
	if err != nil {
		return err
	}

	for _, zone := range config.Config.Deployment.Zones {
		dZone := zone
		log.Println("Setting up log stream for zone", zone.Name)
		go func() {
			err = k8s_service.New().SetupLogStream(ctx, &dZone, allowedNames, func(line string, name string, podNumber int, createdAt time.Time) {
				err = deployment_repo.New().AddLogsByName(name, model.Log{
					Source:    model.LogSourcePod,
					Prefix:    fmt.Sprintf("[pod %d]", podNumber),
					Line:      line,
					CreatedAt: createdAt,
				})
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to add k8s logs for deployment %s. details: %w", name, err))
					return
				}
			})

			if err != nil {
				// TODO: Temporary
				//utils.PrettyPrintError(fmt.Errorf("failed to setup log stream for zone %s. details: %w", zone.Name, err))
				return
			}
		}()
	}

	return nil
}
