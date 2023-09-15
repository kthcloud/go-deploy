package migrator

import (
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s/models"
	"go.mongodb.org/mongo-driver/bson"
	"log"
)

// Migrate will  run as early as possible in the program, and it will never be called again.
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
func getMigrations() map[string]func() error {
	return map[string]func() error{
		"move old port env to new": moveOldPortEnvToNew,
	}
}

func moveOldPortEnvToNew() error {
	deployments, err := deploymentModel.New().GetAll()
	if err != nil {
		return err
	}

	migratedCount := 0

	for _, deployment := range deployments {
		mainApp := deployment.GetMainApp()
		if mainApp == nil {
			return fmt.Errorf("deployment %s has no main app", deployment.Name)
		}

		if mainApp.InternalPort == 0 {
			mainApp.InternalPort = conf.Env.Deployment.Port
		}

		port := mainApp.InternalPort

		k8s := &deployment.Subsystems.K8s

		for name, k8sDeployment := range k8s.DeploymentMap {
			if name == "main" {
				hasNewEnvs := false
				for _, env := range k8sDeployment.EnvVars {
					if env.Name == "PORT" {
						hasNewEnvs = true
						break
					}
				}

				if !hasNewEnvs {
					k8sDeployment.EnvVars = append(k8sDeployment.EnvVars, models.EnvVar{
						Name:  "PORT",
						Value: fmt.Sprintf("%d", port),
					})
				}

				k8s.DeploymentMap[name] = k8sDeployment
			}
		}

		// update subsystems.k8s.deploymentMap and mainApp in array apps
		err = deploymentModel.New().UpdateWithBsonByID(deployment.ID, bson.D{
			{"apps", deployment.Apps},
			{"subsystems.k8s.deploymentMap", k8s.DeploymentMap},
		})
		if err != nil {
			return err
		}

		migratedCount += 1
	}

	log.Println("migratedCount", migratedCount, "/", len(deployments), "deployments")

	return nil
}
