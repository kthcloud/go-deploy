package migrator

import (
	deploymentModels "go-deploy/models/sys/deployment"
	"go.mongodb.org/mongo-driver/bson"
	"log"
)

// Migrate will run as early as possible in the program, and it will never be called again.
func Migrate() {
	migrations := getMigrations()

	if len(migrations) > 0 {
		log.Println("migrating...")

		for name, migration := range migrations {
			log.Printf("running migration %s", name)
			if err := migration(); err != nil {
				log.Fatalf("failed to run migration %s. details: %s", name, err)
			}
		}

		log.Println("migrations done")
		return
	}

	log.Println("nothing to migrate")
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
		"migrateCustomDomainStatusReadyToActive_2024_01_12": migrateCustomDomainStatusReadyToActive_2024_01_12,
	}
}

func migrateCustomDomainStatusReadyToActive_2024_01_12() error {
	deployments, err := deploymentModels.New().List()
	if err != nil {
		return err
	}

	for _, deployment := range deployments {
		app := deployment.GetMainApp()
		if app.CustomDomain != nil && app.CustomDomain.Status == "ready" {
			err = deploymentModels.New().UpdateWithBsonByID(deployment.ID, bson.D{
				{"apps.main.customDomain.status", deploymentModels.CustomDomainStatusActive},
			})

			if err != nil {
				return err
			}
		}
	}

	return nil
}
