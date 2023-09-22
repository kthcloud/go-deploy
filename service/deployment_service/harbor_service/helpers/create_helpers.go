package helpers

import (
	"errors"
	harborModels "go-deploy/pkg/subsystems/harbor/models"
)

func (client *Client) CreateProject(id string, public *harborModels.ProjectPublic) (*harborModels.ProjectPublic, error) {
	createdID, err := client.SsClient.CreateProject(public)
	if err != nil {
		return nil, err
	}

	project, err := client.SsClient.ReadProject(createdID)
	if err != nil {
		return nil, err
	}

	if project == nil {
		return nil, errors.New("failed to read project after creation")
	}

	err = client.UpdateDB(id, "project", project)
	if err != nil {
		return nil, err
	}

	client.SS.Project = *project

	return project, nil
}

func (client *Client) CreateRobot(id string, public *harborModels.RobotPublic) (*harborModels.RobotPublic, error) {
	createdID, err := client.SsClient.CreateRobot(public)
	if err != nil {
		return nil, err
	}

	robot, err := client.SsClient.ReadRobot(createdID)
	if err != nil {
		return nil, err
	}

	if robot == nil {
		return nil, errors.New("failed to read robot after creation")
	}

	err = client.UpdateDB(id, "robot", robot)
	if err != nil {
		return nil, err
	}

	client.SS.Robot = *robot

	return robot, nil
}

func (client *Client) CreateRepository(id string, public *harborModels.RepositoryPublic) (*harborModels.RepositoryPublic, error) {
	_, err := client.SsClient.CreateRepository(public)
	if err != nil {
		return nil, err
	}

	repository, err := client.SsClient.ReadRepository(public.ProjectName, public.Name)
	if err != nil {
		return nil, err
	}

	if repository == nil {
		return nil, errors.New("failed to read repository after creation")
	}

	err = client.UpdateDB(id, "repository", repository)
	if err != nil {
		return nil, err
	}

	client.SS.Repository = *repository

	return repository, nil
}

func (client *Client) CreateWebhook(id string, public *harborModels.WebhookPublic) (*harborModels.WebhookPublic, error) {
	createdID, err := client.SsClient.CreateWebhook(public)
	if err != nil {
		return nil, err
	}

	webhook, err := client.SsClient.ReadWebhook(public.ProjectID, createdID)
	if err != nil {
		return nil, err
	}

	if webhook == nil {
		return nil, errors.New("failed to read webhook after creation")
	}

	err = client.UpdateDB(id, "webhook", webhook)
	if err != nil {
		return nil, err
	}

	client.SS.Webhook = *webhook

	return webhook, nil
}

func (client *Client) CreatePlaceholder(id string) error {
	err := client.UpdateDB(id, "placeholder", true)
	if err != nil {
		return err
	}

	client.SS.Placeholder = true

	return nil
}
