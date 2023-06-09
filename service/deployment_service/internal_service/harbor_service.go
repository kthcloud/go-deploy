package internal_service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/harbor"
	harborModels "go-deploy/pkg/subsystems/harbor/models"
	"go-deploy/utils/subsystemutils"
	"log"
	"reflect"
)

func createProjectPublic(projectName string) *harborModels.ProjectPublic {
	return &harborModels.ProjectPublic{
		Name: projectName,
	}
}

func createRobotPublic(name string, projectID int, projectName string) *harborModels.RobotPublic {
	return &harborModels.RobotPublic{
		Name:        name,
		ProjectID:   projectID,
		ProjectName: projectName,
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
			ProjectName:    conf.Env.DockerRegistry.Placeholder.Project,
			RepositoryName: conf.Env.DockerRegistry.Placeholder.Repository,
		},
	}
}

func createWebhookPublic(projectID int, projectName string) *harborModels.WebhookPublic {
	webhookTarget := fmt.Sprintf("%s/v1/hooks/deployments/harbor", conf.Env.ExternalUrl)
	return &harborModels.WebhookPublic{
		Name:        uuid.NewString(),
		ProjectID:   projectID,
		ProjectName: projectName,
		Target:      webhookTarget,
		Token:       conf.Env.Harbor.WebhookSecret,
	}
}

func withHarborClient() (*harbor.Client, error) {
	return harbor.New(&harbor.ClientConf{
		ApiUrl:   conf.Env.Harbor.Url,
		Username: conf.Env.Harbor.User,
		Password: conf.Env.Harbor.Password,
	})
}

func CreateHarbor(name, userID string) error {
	log.Println("setting up harbor for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup harbor for deployment %s. details: %s", name, err)
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
		log.Println("deployment", name, "not found for harbor setup. assuming it was deleted")
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
		robot, err = createRobot(client, deployment, createRobotPublic(name, project.ID, project.Name))
		if err != nil {
			return makeError(err)
		}
	}

	// Repository
	repository := &deployment.Subsystems.Harbor.Repository
	if !repository.Created() {
		repository, err = createRepository(client, deployment, createRepositoryPublic(project.ID, project.Name, name))
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

func DeleteHarbor(name string) error {
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

func RepairHarbor(name string) error {
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

func recreateProject(client *harbor.Client, deployment *deploymentModel.Deployment, public *harborModels.ProjectPublic) error {
	// we need to delete sub-objects first
	// otherwise harbor will complain about the project still having sub-objects

	id := deployment.Subsystems.Harbor.Project.ID
	name := deployment.Subsystems.Harbor.Project.Name

	err := client.DeleteAllRobots(id)
	if err != nil {
		return err
	}

	err = client.DeleteAllRepositories(name)
	if err != nil {
		return err
	}

	err = client.DeleteAllWebhooks(id)
	if err != nil {
		return err
	}

	err = client.DeleteProject(deployment.Subsystems.Harbor.Project.ID)
	if err != nil {
		return err
	}

	_, err = createProject(client, deployment, public)
	if err != nil {
		return err
	}

	return nil
}

func recreateRobot(client *harbor.Client, deployment *deploymentModel.Deployment, public *harborModels.RobotPublic) error {
	err := client.DeleteRobot(deployment.Subsystems.Harbor.Robot.ID)
	if err != nil {
		return err
	}

	_, err = createRobot(client, deployment, public)
	if err != nil {
		return err
	}

	return nil
}

func recreateRepository(client *harbor.Client, deployment *deploymentModel.Deployment, public *harborModels.RepositoryPublic) error {
	err := client.DeleteRepository(deployment.Subsystems.Harbor.Repository.ProjectName, deployment.Subsystems.Harbor.Repository.Name)
	if err != nil {
		return err
	}

	_, err = createRepository(client, deployment, public)
	if err != nil {
		return err
	}

	return nil
}

func recreateWebhook(client *harbor.Client, deployment *deploymentModel.Deployment, public *harborModels.WebhookPublic) error {
	err := client.DeleteWebhook(deployment.Subsystems.Harbor.Webhook.ProjectID, deployment.Subsystems.Harbor.Webhook.ID)
	if err != nil {
		return err
	}

	_, err = createWebhook(client, deployment, public)
	if err != nil {
		return err
	}

	return nil
}

func createProject(client *harbor.Client, deployment *deploymentModel.Deployment, public *harborModels.ProjectPublic) (*harborModels.ProjectPublic, error) {
	id, err := client.CreateProject(public)
	if err != nil {
		return nil, err
	}

	project, err := client.ReadProject(id)
	if err != nil {
		return nil, err
	}

	if project == nil {
		return nil, errors.New("failed to read project after creation")
	}

	err = deploymentModel.UpdateSubsystemByName(deployment.Name, "harbor", "project", project)
	if err != nil {
		return nil, err
	}

	deployment.Subsystems.Harbor.Project = *project

	return project, nil
}

func createRobot(client *harbor.Client, deployment *deploymentModel.Deployment, public *harborModels.RobotPublic) (*harborModels.RobotPublic, error) {
	id, err := client.CreateRobot(public)
	if err != nil {
		return nil, err
	}

	robot, err := client.ReadRobot(id)
	if err != nil {
		return nil, err
	}

	if robot == nil {
		return nil, errors.New("failed to read robot after creation")
	}

	err = deploymentModel.UpdateSubsystemByName(deployment.Name, "harbor", "robot", robot)
	if err != nil {
		return nil, err
	}

	deployment.Subsystems.Harbor.Robot = *robot

	return robot, nil
}

func createRepository(client *harbor.Client, deployment *deploymentModel.Deployment, public *harborModels.RepositoryPublic) (*harborModels.RepositoryPublic, error) {
	_, err := client.CreateRepository(public)
	if err != nil {
		return nil, err
	}

	repository, err := client.ReadRepository(public.ProjectName, public.Name)
	if err != nil {
		return nil, err
	}

	if repository == nil {
		return nil, errors.New("failed to read repository after creation")
	}

	err = deploymentModel.UpdateSubsystemByName(deployment.Name, "harbor", "repository", repository)
	if err != nil {
		return nil, err
	}

	deployment.Subsystems.Harbor.Repository = *repository

	return repository, nil
}

func createWebhook(client *harbor.Client, deployment *deploymentModel.Deployment, public *harborModels.WebhookPublic) (*harborModels.WebhookPublic, error) {
	id, err := client.CreateWebhook(public)
	if err != nil {
		return nil, err
	}

	webhook, err := client.ReadWebhook(public.ProjectID, id)
	if err != nil {
		return nil, err
	}

	if webhook == nil {
		return nil, errors.New("failed to read webhook after creation")
	}

	err = deploymentModel.UpdateSubsystemByName(deployment.Name, "harbor", "webhook", webhook)
	if err != nil {
		return nil, err
	}

	deployment.Subsystems.Harbor.Webhook = *webhook

	return webhook, nil
}
