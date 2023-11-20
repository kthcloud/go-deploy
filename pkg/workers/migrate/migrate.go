package migrator

import (
	deploymentModel "go-deploy/models/sys/deployment"
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
		"2023-11-10_add_replicas_to_main_app": addReplicasToMainApp,
	}
}

func addReplicasToMainApp() error {
	deployments, err := deploymentModel.New().List()
	if err != nil {
		return err
	}

	for _, deployment := range deployments {
		app := deployment.GetMainApp()
		if app == nil {
			continue
		}

		if app.Replicas == 0 {
			app.Replicas = 1

			update := bson.D{
				{"apps.main", app},
			}

			err = deploymentModel.New().SetWithBsonByID(deployment.ID, update)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
