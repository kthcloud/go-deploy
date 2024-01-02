package migrator

import (
	"go-deploy/models/sys/base/resource"
	deploymentModels "go-deploy/models/sys/deployment"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"strings"
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
		"migrateOldCustomDomainToNewStruct_2024_01_02": migrateOldCustomDomainToNewStruct_2024_01_02,
	}
}

type oldApp struct {
	CustomDomain *string `bson:"customDomain"`
}

type oldDeployment struct {
	ID   string            `bson:"id"`
	Apps map[string]oldApp `bson:"apps"`
}

func migrateOldCustomDomainToNewStruct_2024_01_02() error {
	rc := resource.ResourceClient[oldDeployment]{
		Collection: deploymentModels.New().Collection,
	}

	ids, err := deploymentModels.New().ListIDs()
	if err != nil {
		return err
	}

	for _, id := range ids {
		deployment, err := rc.GetByID(id.ID)
		if err != nil {
			if strings.Contains(err.Error(), "error decoding key") {
				continue
			}

			return err
		}

		mainApp := deployment.Apps["main"]
		if mainApp.CustomDomain != nil {
			cd := deploymentModels.CustomDomain{
				Domain: *mainApp.CustomDomain,
				Secret: "",
				Status: deploymentModels.CustomDomainStatusReady,
			}

			update := bson.D{
				{"apps.main.customDomain", cd},
			}

			err = deploymentModels.New().SetWithBsonByID(id.ID, update)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
