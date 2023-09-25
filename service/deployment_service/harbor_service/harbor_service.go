package harbor_service

import (
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/pkg/subsystems/harbor/models"
	"go-deploy/service"
	"go-deploy/service/deployment_service/harbor_service/helpers"
	"go-deploy/utils/subsystemutils"
	"log"
)

func Create(deploymentID, userID string, params *deploymentModel.CreateParams) error {
	log.Println("setting up harbor for", params.Name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup harbor for deployment %s. details: %w", params.Name, err)
	}

	deployment, err := deploymentModel.New().GetByID(deploymentID)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", deploymentID, "not found for harbor setup. assuming it was deleted")
		return nil
	}

	client, err := helpers.New(&deployment.Subsystems.Harbor)
	if err != nil {
		return makeError(err)
	}

	// Project
	project := &deployment.Subsystems.Harbor.Project
	if service.NotCreated(project) {
		project, err = client.CreateProject(deployment.ID, helpers.CreateProjectPublic(subsystemutils.GetPrefixedName(userID)))
		if err != nil {
			return makeError(err)
		}
	}

	// Robot
	if service.NotCreated(&deployment.Subsystems.Harbor.Robot) {
		_, err = client.CreateRobot(deployment.ID, helpers.CreateRobotPublic(deployment.Name, project.ID, project.Name))
		if err != nil {
			return makeError(err)
		}
	}

	// Repository
	if service.NotCreated(&deployment.Subsystems.Harbor.Repository) {
		_, err = client.CreateRepository(deployment.ID, helpers.CreateRepositoryPublic(project.ID, project.Name, deployment.Name))
		if err != nil {
			return makeError(err)
		}
	}

	// Webhook
	if service.NotCreated(&deployment.Subsystems.Harbor.Webhook) {
		_, err = client.CreateWebhook(deployment.ID, helpers.CreateWebhookPublic(project.ID, project.Name))
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func CreatePlaceholder(name string) error {
	log.Println("setting up placeholder harbor")

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup placeholder github. details: %w", err)
	}

	deployment, err := deploymentModel.New().GetByName(name)
	if err != nil {
		return makeError(err)
	}

	client, err := helpers.New(&deployment.Subsystems.Harbor)
	if err != nil {
		return makeError(err)
	}

	err = client.CreatePlaceholder(deployment.ID)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Delete(name string) error {
	log.Println("deleting harbor for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete harbor for deployment %s. details: %w", name, err)
	}

	deployment, err := deploymentModel.New().GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", name, "not found for harbor deletion. assuming it was deleted")
		return nil
	}

	client, err := helpers.New(&deployment.Subsystems.Harbor)
	if err != nil {
		return makeError(err)
	}

	err = client.DeleteRepository(deployment.ID)
	if err != nil {
		return makeError(err)
	}

	err = client.DeleteRobot(deployment.ID)
	if err != nil {
		return makeError(err)
	}

	err = client.DeleteProject(deployment.ID)
	if err != nil {
		return makeError(err)
	}

	err = client.DeleteWebhook(deployment.ID)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Repair(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair harbor %s. details: %w", name, err)
	}

	deployment, err := deploymentModel.New().GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", name, "not found when repairing harbor, assuming it was deleted")
		return nil
	}

	client, err := helpers.New(&deployment.Subsystems.Harbor)
	if err != nil {
		return makeError(err)
	}

	err = client.RepairProject(deployment.ID, func() *models.ProjectPublic {
		return helpers.CreateProjectPublic(subsystemutils.GetPrefixedName(deployment.OwnerID))
	})

	err = client.RepairRobot(deployment.ID, func() *models.RobotPublic {
		return helpers.CreateRobotPublic(deployment.Name, client.SS.Project.ID, client.SS.Project.Name)
	})

	err = client.RepairRepository(deployment.ID, func() *models.RepositoryPublic {
		return helpers.CreateRepositoryPublic(client.SS.Project.ID, client.SS.Project.Name, deployment.Name)
	})

	err = client.RepairWebhook(deployment.ID, func() *models.WebhookPublic {
		return helpers.CreateWebhookPublic(client.SS.Project.ID, client.SS.Project.Name)
	})

	return nil
}
