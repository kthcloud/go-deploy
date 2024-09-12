package migrator

import (
	"fmt"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/deployment_repo"
	"github.com/kthcloud/go-deploy/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
)

// Migrate runs every migration specified in the getMigrations function.
// It should be run as early as possible in the program, and should never be called again.
func Migrate() error {
	migrations := getMigrations()

	if len(migrations) > 0 {
		for name, migration := range migrations {
			log.Printf("- %s (%d/%d)", name, 1, len(migrations))
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
		"migratePrivateBooleanToVisibilityEnum_2024_06_10": migratePrivateBooleanToVisibilityEnum_2024_06_10,
	}
}

func migratePrivateBooleanToVisibilityEnum_2024_06_10() error {
	deployments, err := deployment_repo.New().List()
	if err != nil {
		return err
	}

	for _, deployment := range deployments {
		mainApp := deployment.GetMainApp()

		if mainApp.Visibility != "" {
			continue
		}

		if mainApp.Private {
			mainApp.Visibility = model.VisibilityPrivate
		} else {
			mainApp.Visibility = model.VisibilityPublic
		}

		err = deployment_repo.New().SetWithBsonByID(deployment.ID, bson.D{{"apps.main.visibility", mainApp.Visibility}})
		if err != nil {
			return err
		}
	}

	return nil
}
