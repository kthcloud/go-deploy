package deployment_service

import (
	"fmt"
	deploymentModel "go-deploy/models/deployment"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/harbor"
	harborModels "go-deploy/pkg/subsystems/harbor/models"
	"go-deploy/utils/subsystemutils"
	"log"
)

func createProjectPublic(projectName string) *harborModels.ProjectPublic {
	return &harborModels.ProjectPublic{
		Name:   projectName,
		Public: true,
	}
}

func createRobotPublic(projectID int, projectName, name string) *harborModels.RobotPublic {
	return &harborModels.RobotPublic{
		Name:        name,
		ProjectID:   projectID,
		ProjectName: projectName,
		Description: "Auto created with Go Deploy",
		Disable:     false,
	}
}

func createRepositoryPublic(projectID int, projectName string, name string) *harborModels.RepositoryPublic {
	return &harborModels.RepositoryPublic{
		Name:        name,
		ProjectID:   projectID,
		ProjectName: projectName,
		Seeded:      false,
		Placeholder: &harborModels.PlaceHolder{
			ProjectName:    conf.Env.DockerRegistry.PlaceHolderProject,
			RepositoryName: conf.Env.DockerRegistry.PlaceHolderRepository,
		},
	}
}

func createWebhookPublic(projectID int, projectName, name string) *harborModels.WebhookPublic {
	webhookTarget := fmt.Sprintf("%s/v1/hooks/deployments/harbor", conf.Env.ExternalUrl)
	return &harborModels.WebhookPublic{
		Name:        name,
		ProjectID:   projectID,
		ProjectName: projectName,
		Target:      webhookTarget,
		Token:       conf.Env.Harbor.WebhookSecret,
	}
}

func CreateHarbor(name string) error {
	log.Println("setting up harbor for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup harbor for v1_deployment %s. details: %s", name, err)
	}

	client, err := harbor.New(&harbor.ClientConf{
		ApiUrl:   conf.Env.Harbor.Url,
		Username: conf.Env.Harbor.Identity,
		Password: conf.Env.Harbor.Secret,
	})
	if err != nil {
		return makeError(err)
	}

	deployment, err := deploymentModel.GetDeploymentByName(name)
	if err != nil {
		return makeError(err)
	}

	// Project
	var project *harborModels.ProjectPublic
	if deployment.Subsystems.Harbor.Project.ID == 0 {
		projectName := subsystemutils.GetPrefixedName(name)
		id, err := client.CreateProject(createProjectPublic(projectName))
		if err != nil {
			return makeError(err)
		}

		project, err = client.ReadProject(id)
		if err != nil {
			return makeError(err)
		}

		err = deploymentModel.UpdateSubsystemByName(name, "harbor", "project", project)
		if err != nil {
			return makeError(err)
		}
	} else {
		project, err = client.ReadProject(deployment.Subsystems.Harbor.Project.ID)
		if err != nil {
			return makeError(err)
		}
	}

	// Robot
	if deployment.Subsystems.Harbor.Robot.ID == 0 {
		created, err := client.CreateRobot(createRobotPublic(project.ID, project.Name, name))
		if err != nil {
			return makeError(err)
		}

		robot, err := client.ReadRobot(created.ID)
		if err != nil {
			return makeError(err)
		}

		robot.Secret = created.Secret

		err = deploymentModel.UpdateSubsystemByName(name, "harbor", "robot", robot)
		if err != nil {
			return makeError(err)
		}
	}

	// Repository
	if deployment.Subsystems.Harbor.Repository.ID == 0 {
		_, err := client.CreateRepository(createRepositoryPublic(project.ID, project.Name, name))
		if err != nil {
			return makeError(err)
		}

		repository, err := client.ReadRepository(project.Name, name)
		if err != nil {
			return makeError(err)
		}

		err = deploymentModel.UpdateSubsystemByName(name, "harbor", "repository", repository)
		if err != nil {
			return makeError(err)
		}
	}

	// Webhook
	if deployment.Subsystems.Harbor.Webhook.ID == 0 {
		id, err := client.CreateWebhook(createWebhookPublic(project.ID, project.Name, name))
		if err != nil {
			return makeError(err)
		}

		webhook, err := client.ReadWebhook(project.ID, id)
		if err != nil {
			return makeError(err)
		}

		err = deploymentModel.UpdateSubsystemByName(name, "harbor", "webhook", webhook)
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func DeleteHarbor(name string) error {
	log.Println("deleting harbor setup for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete harbor setup for v1_deployment %s. details: %s", name, err)
	}

	client, err := harbor.New(&harbor.ClientConf{
		ApiUrl:   conf.Env.Harbor.Url,
		Username: conf.Env.Harbor.Identity,
		Password: conf.Env.Harbor.Secret,
	})
	if err != nil {
		return makeError(err)
	}

	deployment, err := deploymentModel.GetDeploymentByName(name)

	if deployment.Subsystems.Harbor.Webhook.ID != 0 {
		err = deploymentModel.UpdateSubsystemByName(name, "harbor", "webhook", harborModels.WebhookPublic{})
		if err != nil {
			return makeError(err)
		}
	}

	if deployment.Subsystems.Harbor.Repository.ID != 0 {
		err = client.DeleteRepository(deployment.Subsystems.Harbor.Repository.ProjectName, deployment.Subsystems.Harbor.Repository.Name)
		if err != nil {
			return makeError(err)
		}

		err = deploymentModel.UpdateSubsystemByName(name, "harbor", "repository", harborModels.RepositoryPublic{})
		if err != nil {
			return makeError(err)
		}
	}

	if deployment.Subsystems.Harbor.Robot.ID != 0 {
		err = deploymentModel.UpdateSubsystemByName(name, "harbor", "robot", harborModels.RobotPublic{})
		if err != nil {
			return makeError(err)
		}
	}

	if deployment.Subsystems.Harbor.Project.ID != 0 {
		err = client.DeleteProject(deployment.Subsystems.Harbor.Project.ID)
		if err != nil {
			return makeError(err)
		}

		err = deploymentModel.UpdateSubsystemByName(name, "harbor", "project", harborModels.ProjectPublic{})
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}
