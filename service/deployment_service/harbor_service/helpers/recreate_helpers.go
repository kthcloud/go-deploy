package helpers

import (
	harborModels "go-deploy/pkg/subsystems/harbor/models"
)

func (client *Client) RecreateProject(id string, newPublic *harborModels.ProjectPublic) error {
	// we need to delete sub-objects first
	// otherwise harbor will complain about the project still having sub-objects

	projectID := client.SS.Project.ID
	projectName := client.SS.Project.Name

	err := client.SsClient.DeleteAllRobots(projectID)
	if err != nil {
		return err
	}

	err = client.SsClient.DeleteAllRepositories(projectName)
	if err != nil {
		return err
	}

	err = client.SsClient.DeleteAllWebhooks(projectID)
	if err != nil {
		return err
	}

	err = client.SsClient.DeleteProject(projectID)
	if err != nil {
		return err
	}

	_, err = client.CreateProject(id, newPublic)
	return err
}

func (client *Client) RecreateRobot(id string, newPublic *harborModels.RobotPublic) error {
	err := client.DeleteRobot(id)
	if err != nil {
		return err
	}

	_, err = client.CreateRobot(id, newPublic)
	return err
}

func (client *Client) RecreateRepository(id string, newPublic *harborModels.RepositoryPublic) error {
	err := client.DeleteRepository(id)
	if err != nil {
		return err
	}

	_, err = client.CreateRepository(id, newPublic)
	return err
}

func (client *Client) RecreateWebhook(id string, newPublic *harborModels.WebhookPublic) error {
	err := client.DeleteWebhook(id)
	if err != nil {
		return err
	}

	_, err = client.CreateWebhook(id, newPublic)
	return err
}
