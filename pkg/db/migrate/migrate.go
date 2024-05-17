package migrator

import (
	"fmt"
	"go-deploy/pkg/db/resources/deployment_repo"
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
		if deployment.AccessedAt.IsZero() {
			// Set to the greatest of:
			//UpdatedAt   time.Time `bson:"updatedAt"`
			//RepairedAt  time.Time `bson:"repairedAt"`
			//RestartedAt time.Time `bson:"restartedAt"`

			// Find the greatest time
			greatestTime := deployment.UpdatedAt
			if deployment.RepairedAt.After(greatestTime) {
				greatestTime = deployment.RepairedAt
			}

			if deployment.RestartedAt.After(greatestTime) {
				greatestTime = deployment.RestartedAt
			}

			deployment.AccessedAt = greatestTime

			err = deployment_repo.New().SetWithBsonByID(deployment.ID, bson.D{{"accessedAt", deployment.AccessedAt}})
			if err != nil {
				return err
			}
		}
	}

	for _, vm := range vms {
		if vm.AccessedAt.IsZero() {
			// Find the greatest time
			greatestTime := vm.UpdatedAt
			if vm.RepairedAt.After(greatestTime) {
				greatestTime = vm.RepairedAt
			}

			vm.AccessedAt = greatestTime

			err = vm_repo.New().SetWithBsonByID(vm.ID, bson.D{{"accessedAt", vm.AccessedAt}})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
