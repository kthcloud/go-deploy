package migrator

import (
	"fmt"
	"go-deploy/pkg/db/resources/sm_repo"
	"go-deploy/pkg/log"
	v1 "go-deploy/service/v1"
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
		"migrateSmsToNewSeFlemZone_2024_04_23": migrateSmsToNewSeFlemZone_2024_04_23,
	}
}

func migrateSmsToNewSeFlemZone_2024_04_23() error {
	sms, err := sm_repo.New().WithZone("se-flem").List()
	if err != nil {
		return err
	}

	for _, sm := range sms {
		sm.Zone = "se-flem-2"
		if err := sm_repo.New().SetWithBsonByID(sm.ID, bson.D{{"zone", sm.Zone}}); err != nil {
			return err
		}

		// Repair
		if err := v1.New().SMs().Repair(sm.ID); err != nil {
			// Revert zone change
			_ = sm_repo.New().SetWithBsonByID(sm.ID, bson.D{{"zone", "se-flem"}})
			return err
		}
	}

	return nil
}
