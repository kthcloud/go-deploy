package migrator

import (
	"fmt"
	"go-deploy/pkg/db/resources/deployment_repo"
	"go-deploy/pkg/log"
	"go-deploy/service"
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
		"moveToNewZoneSeFlem_2024_04_23": moveToNewZoneSeFlem_2024_04_23,
	}
}

func moveToNewZoneSeFlem_2024_04_23() error {
	deployments, err := deployment_repo.New().List()
	if err != nil {
		return err
	}

	for _, deployment := range deployments {
		if err := service.V1().Deployments().Repair(deployment.ID); err != nil {
			return err
		}
	}

	return nil
}
