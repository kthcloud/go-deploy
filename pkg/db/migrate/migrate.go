package migrator

import (
	"fmt"
	"go-deploy/pkg/db/resources/deployment_repo"
	"go-deploy/pkg/db/resources/user_repo"
	"go-deploy/pkg/db/resources/vm_repo"
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
		"addAccessedAt_2024_05_17": addAccessedAt_2024_05_17,
	}
}

func addAccessedAt_2024_05_17() error {
	deployments, err := deployment_repo.New().List()
	if err != nil {
		return err
	}

	vms, err := vm_repo.New().List()
	if err != nil {
		return err
	}

	for _, deployment := range deployments {
		if deployment.OwnerID == "system" {
			continue
		}

		user, err := user_repo.New().GetByID(deployment.OwnerID)
		if err != nil {
			return err
		}

		if user == nil {
			return fmt.Errorf("user not found for deployment %s", deployment.ID)
		}

		err = deployment_repo.New().SetWithBsonByID(deployment.ID, bson.D{{"accessedAt", user.LastAuthenticatedAt}})
		if err != nil {
			return err
		}
	}

	for _, vm := range vms {
		if vm.OwnerID == "system" {
			continue
		}

		user, err := user_repo.New().GetByID(vm.OwnerID)
		if err != nil {
			return err
		}

		if user == nil {
			return fmt.Errorf("user not found for vm %s", vm.ID)
		}

		err = vm_repo.New().SetWithBsonByID(vm.ID, bson.D{{"accessedAt", user.LastAuthenticatedAt}})
		if err != nil {
			return err
		}
	}

	return nil
}
