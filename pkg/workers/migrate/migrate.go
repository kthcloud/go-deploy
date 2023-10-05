package migrator

import (
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/service/deployment_service"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
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
		"2023-10-05 fix typos in subsystems":              fixTyposInSubsystems_2023_10_05,
		"2022-10-05 fix harbor project public to private": fixHarborProjectPublicToPrivate_2022_10_05,
		"2022-10-05 add image pull secret to deployments": addImagePullSecretToDeployments_2022_10_05,
	}
}

func fixTyposInSubsystems_2023_10_05() error {
	deployments, err := deployment_service.GetAll()
	if err != nil {
		return err
	}

	for _, deployment := range deployments {
		update := bson.D{}

		mainK8sDeployment := deployment.Subsystems.K8s.GetDeployment("main")
		if mainK8sDeployment != nil && mainK8sDeployment.Image == "" {
			update = append(update, bson.E{Key: "subsystems.k8s.deploymentMap.main.image", Value: mainK8sDeployment.DockerImage})
		}

		if deployment.Subsystems.Harbor.Repository.ID != 0 {
			update = append(update, bson.E{Key: "subsystems.harbor.repository.id", Value: deployment.Subsystems.Harbor.Repository.ID})
		}

		if len(update) == 0 {
			continue
		}

		update = append(update, bson.E{Key: "repairAt", Value: time.Time{}})

		err = deploymentModel.New().UpdateWithBsonByID(deployment.ID, update)
		if err != nil {
			return err
		}
	}

	return nil
}

func fixHarborProjectPublicToPrivate_2022_10_05() error {
	deployments, err := deployment_service.GetAll()
	if err != nil {
		return err
	}

	for _, deployment := range deployments {
		if !deployment.Subsystems.Harbor.Project.Public {
			continue
		}

		update := bson.D{}

		update = append(update, bson.E{Key: "subsystems.harbor.project.public", Value: false})
		update = append(update, bson.E{Key: "repairAt", Value: time.Time{}})

		err = deploymentModel.New().UpdateWithBsonByID(deployment.ID, update)
		if err != nil {
			return err
		}
	}

	return nil
}

func addImagePullSecretToDeployments_2022_10_05() error {
	deployments, err := deployment_service.GetAll()
	if err != nil {
		return err
	}

	for _, deployment := range deployments {
		if len(deployment.Subsystems.K8s.GetSecretMap()) > 0 {
			continue
		}

		update := bson.D{}

		imagePullSecrets := []string{
			deployment.Name + "-image-pull-secret",
		}

		update = append(update, bson.E{Key: "subsystems.k8s.deploymentMap.main.imagePullSecrets", Value: imagePullSecrets})
		update = append(update, bson.E{Key: "repairAt", Value: time.Time{}})

		err = deploymentModel.New().UpdateWithBsonByID(deployment.ID, update)
		if err != nil {
			return err
		}
	}

	return nil
}
