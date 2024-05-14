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
)

// deploymentLogger is a worker that logs deployments.
func deploymentLogger(ctx context.Context) error {
	// Get allowed names
	allowedNames, err := deployment_repo.New().ListNames()
	if err != nil {
		return err
	}

	for _, zone := range config.Config.Zones {
		dZone := zone
		log.Println("Setting up log stream for zone", zone.Name)
		go func() {
			err = k8s_service.New().SetupLogStream(ctx, &dZone, allowedNames, func(name string, lines []model.Log) {
				err = deployment_repo.New().AddLogsByName(name, lines...)
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to add k8s logs for deployment %s. details: %w", name, err))
					return
				}
			})

			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to set up log stream for zone %s. details: %w", dZone.Name, err))
				return
			}
		}()
	}

	return nil
}
