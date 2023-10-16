package migrator

import (
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/service"
	"go-deploy/utils/subsystemutils"
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
		"2023-10-16-set-correct-namespace-name": setCorrectNamespaceName_2023_10_16,
	}
}

func setCorrectNamespaceName_2023_10_16() error {
	deployments, err := deploymentModels.New().GetAll()
	if err != nil {
		return err
	}

	for _, deployment := range deployments {
		if ns := deployment.Subsystems.K8s.GetNamespace(); service.Created(ns) {
			correctName := subsystemutils.GetPrefixedName(deployment.OwnerID)

			if ns.Name == correctName {
				continue
			}

			ns.Name = correctName

			err = deploymentModels.New().UpdateSubsystemByID(deployment.ID, "k8s.namespace", ns)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
