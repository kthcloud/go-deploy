package migrator

import (
	"fmt"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/deployment_repo"
	"go-deploy/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
)

// Migrate runs every migration specified in the getMigrations function.
// It should be run as early as possible in the program, and should never be called again.
func Migrate() error {
	migrations := getMigrations()

	if len(migrations) > 0 {
		for name, migration := range migrations {
			log.Printf("- %s (%d/%d)\n", name, 1, len(migrations))
			if err := migration(); err != nil {
				return fmt.Errorf("migration %s failed. details: %w", name, err)
			}
		}
	} else {
		log.Println("No migrations to run")
	}

	return nil
}

// getMigrations returns a map of migrations to run.
// add a migration to the list of functions to run.
// clear when prod has run it once.
//
// the migrations must be **idempotent**.
//
// add a date to the migration name to make it easier to identify.
func getMigrations() map[string]func() error {
	return map[string]func() error{
		"addSpecsToDeploymentWithoutSpecs_2024_05_10": addSpecsToDeploymentWithoutSpecs_2024_05_10,
	}
}

func addSpecsToDeploymentWithoutSpecs_2024_05_10() error {
	deployments, err := deployment_repo.New().List()
	if err != nil {
		return err
	}

	for _, deployment := range deployments {
		for _, app := range deployment.Apps {
			anyUpdated := false

			if app.CpuCores == 0 {
				app.CpuCores = config.Config.Deployment.Resources.Limits.CPU
				anyUpdated = true
			}
			if app.RAM == 0 {
				app.RAM = config.Config.Deployment.Resources.Limits.RAM
				anyUpdated = true
			}

			if anyUpdated {
				err = deployment_repo.New().SetWithBsonByID(deployment.ID, bson.D{{"apps." + app.Name, app}})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
