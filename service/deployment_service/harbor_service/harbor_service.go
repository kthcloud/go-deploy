package harbor_service

import (
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	harborModels "go-deploy/pkg/subsystems/harbor/models"
	"go-deploy/utils/subsystemutils"
	"log"
	"reflect"
)

func Create(deploymentID, userID string, params *deploymentModel.CreateParams) error {
	log.Println("setting up harbor for", params.Name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup harbor for deployment %s. details: %s", params.Name, err)
	}

	client, err := withHarborClient()
	if err != nil {
		return makeError(err)
	}

	deployment, err := deploymentModel.GetByID(deploymentID)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", deploymentID, "not found for harbor setup. assuming it was deleted")
		return nil
	}

	// Project
	project := &deployment.Subsystems.Harbor.Project
	if !project.Created() {
		project, err = createProject(client, deployment, createProjectPublic(subsystemutils.GetPrefixedName(userID)))
		if err != nil {
			return makeError(err)
		}
	}

	// Robot
	robot := &deployment.Subsystems.Harbor.Robot
	if !robot.Created() {
		robot, err = createRobot(client, deployment, createRobotPublic(deployment.Name, project.ID, project.Name))
		if err != nil {
			return makeError(err)
		}
	}

	// Repository
	repository := &deployment.Subsystems.Harbor.Repository
	if !repository.Created() {
		repository, err = createRepository(client, deployment, createRepositoryPublic(project.ID, project.Name, deployment.Name))
		if err != nil {
			return makeError(err)
		}
	}

	// Webhook
	webhook := &deployment.Subsystems.Harbor.Webhook
	if !webhook.Created() {
		webhook, err = createWebhook(client, deployment, createWebhookPublic(project.ID, project.Name))
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func Delete(name string) error {
	log.Println("deleting harbor for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete harbor for deployment %s. details: %s", name, err)
	}

	client, err := withHarborClient()
	if err != nil {
		return makeError(err)
	}

	deployment, err := deploymentModel.GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", name, "not found for harbor deletion. assuming it was deleted")
		return nil
	}

	if deployment.Subsystems.Harbor.Repository.Created() {
		err = client.DeleteRepository(deployment.Subsystems.Harbor.Repository.ProjectName, deployment.Subsystems.Harbor.Repository.Name)
		if err != nil {
			return makeError(err)
		}

		err = deploymentModel.UpdateSubsystemByName(name, "harbor", "repository", harborModels.RepositoryPublic{})
		if err != nil {
			return makeError(err)
		}
	}

	if deployment.Subsystems.Harbor.Robot.Created() {
		err = deploymentModel.UpdateSubsystemByName(name, "harbor", "robot", harborModels.RobotPublic{})
		if err != nil {
			return makeError(err)
		}
	}

	if deployment.Subsystems.Harbor.Webhook.Created() {
		err = deploymentModel.UpdateSubsystemByName(name, "harbor", "webhook", harborModels.WebhookPublic{})
		if err != nil {
			return makeError(err)
		}
	}

	if deployment.Subsystems.Harbor.Project.Created() {
		err = deploymentModel.UpdateSubsystemByName(name, "harbor", "project", harborModels.ProjectPublic{})
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func Repair(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair harbor %s. details: %s", name, err)
	}

	deployment, err := deploymentModel.GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", name, "not found when repairing harbor, assuming it was deleted")
		return nil
	}

	client, err := withHarborClient()
	if err != nil {
		return makeError(err)
	}

	ss := deployment.Subsystems.Harbor

	// project
	project, err := client.ReadProject(ss.Project.ID)
	if err != nil {
		return makeError(err)
	}

	if project == nil || !reflect.DeepEqual(ss.Project, *project) {
		err = recreateProject(client, deployment, &ss.Project)
		if err != nil {
			return makeError(err)
		}
	}

	// robot
	robot, err := client.ReadRobot(ss.Robot.ID)
	if err != nil {
		return makeError(err)
	}

	if robot == nil || !reflect.DeepEqual(ss.Robot, *robot) {
		// reset the secret to not cause confusion, since it won't be used anyway
		ss.Robot.Secret = ""

		err = recreateRobot(client, deployment, &ss.Robot)
		if err != nil {
			return makeError(err)
		}
	}

	// repository
	repository, err := client.ReadRepository(ss.Repository.ProjectName, ss.Repository.Name)
	if err != nil {
		return makeError(err)
	}

	if repository == nil || !reflect.DeepEqual(ss.Repository, *repository) {
		err = recreateRepository(client, deployment, &ss.Repository)
		if err != nil {
			return makeError(err)
		}
	}

	// webhook
	webhook, err := client.ReadWebhook(ss.Webhook.ProjectID, ss.Webhook.ID)
	if err != nil {
		return makeError(err)
	}

	if webhook == nil || !reflect.DeepEqual(ss.Webhook, *webhook) {
		err = recreateWebhook(client, deployment, &ss.Webhook)
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}
