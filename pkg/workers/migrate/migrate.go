package migrator

import (
	deploymentModel "go-deploy/models/sys/deployment"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service/deployment_service"
	"go.mongodb.org/mongo-driver/bson"
	"log"
)

// This file is edited when an update in the database schema has occurred.
// Thus, every migration needed will be done programmatically.
// Once a migration is done, clear the file.

// Migrate will  run as early as possible in the program, and it will never be called again.
func Migrate() {
	migrateToK8sMaps()
	migrateToDeploymentApps()
}

func migrateToK8sMaps() {
	log.Println("migrating deployments to new maps")

	allDeployments, err := deployment_service.GetAll()
	if err != nil {
		panic(err)
	}

	noMigrated := 0

	for _, deployment := range allDeployments {
		oldK8sDeployment := deployment.Subsystems.K8s.Deployment
		oldK8sService := deployment.Subsystems.K8s.Service
		oldK8sIngress := deployment.Subsystems.K8s.Ingress

		if deployment.Subsystems.K8s.DeploymentMap == nil {
			deployment.Subsystems.K8s.DeploymentMap = make(map[string]k8sModels.DeploymentPublic)
		}

		if deployment.Subsystems.K8s.ServiceMap == nil {
			deployment.Subsystems.K8s.ServiceMap = make(map[string]k8sModels.ServicePublic)
		}

		if deployment.Subsystems.K8s.IngressMap == nil {
			deployment.Subsystems.K8s.IngressMap = make(map[string]k8sModels.IngressPublic)
		}

		newK8sDeployment := deployment.Subsystems.K8s.DeploymentMap["main"]
		newK8sService := deployment.Subsystems.K8s.ServiceMap["main"]
		newK8sIngress := deployment.Subsystems.K8s.IngressMap["main"]

		migrated := false

		if !newK8sDeployment.Created() {
			deployment.Subsystems.K8s.DeploymentMap["main"] = oldK8sDeployment
			migrated = true
		}

		if !newK8sService.Created() {
			deployment.Subsystems.K8s.ServiceMap["main"] = oldK8sService
			migrated = true
		}

		if !newK8sIngress.Created() {
			deployment.Subsystems.K8s.IngressMap["main"] = oldK8sIngress
			migrated = true
		}

		if migrated {
			err := deploymentModel.UpdateByID(deployment.ID, bson.D{
				{"subsystems.k8s.deploymentMap", deployment.Subsystems.K8s.DeploymentMap},
			})
			if err != nil {
				panic(err)
			}
			err = deploymentModel.UpdateByID(deployment.ID, bson.D{
				{"subsystems.k8s.serviceMap", deployment.Subsystems.K8s.ServiceMap},
			})
			if err != nil {
				panic(err)
			}
			err = deploymentModel.UpdateByID(deployment.ID, bson.D{
				{"subsystems.k8s.ingressMap", deployment.Subsystems.K8s.IngressMap},
			})
			if err != nil {
				panic(err)
			}

			noMigrated++
		}
	}

	log.Println("migrated", noMigrated, "deployments")
}

func migrateToDeploymentApps() {
	log.Println("migrating deployments to new apps")

	allDeployments, err := deployment_service.GetAll()
	if err != nil {
		panic(err)
	}

	noMigrated := 0
	for _, deployment := range allDeployments {
		if deployment.Apps == nil {
			deployment.Apps = make(map[string]deploymentModel.App)
		}

		mainApp := deployment.GetMainApp()
		if mainApp != nil {
			continue
		}

		deployment.Apps["main"] = deploymentModel.App{
			Name:         "main",
			Private:      deployment.Private,
			Envs:         deployment.Envs,
			Volumes:      deployment.Volumes,
			InitCommands: deployment.InitCommands,
		}

		err = deploymentModel.UpdateByID(deployment.ID, bson.D{
			{"apps", deployment.Apps},
		})
		if err != nil {
			panic(err)
		}

		noMigrated++
	}

	log.Println("migrated", noMigrated, "deployments")
}
