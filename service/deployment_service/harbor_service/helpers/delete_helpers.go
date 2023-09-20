package helpers

import (
	harborModels "go-deploy/pkg/subsystems/harbor/models"
	"go-deploy/service"
)

func (client *Client) DeleteProject(id string) error {
	// since project is shared, we never actually delete it in harbor

	err := client.UpdateDB(id, "project", nil)
	if err != nil {
		return err
	}

	client.SS.Project = harborModels.ProjectPublic{}

	return nil
}

func (client *Client) DeleteRobot(id string) error {
	robot := client.SS.Robot
	if service.Created(&robot) {
		err := client.SsClient.DeleteRobot(robot.ID)
		if err != nil {
			return err
		}
	}

	client.SS.Robot = harborModels.RobotPublic{}
	return client.UpdateDB(id, "robot", harborModels.RobotPublic{})
}

func (client *Client) DeleteRepository(id string) error {
	repository := client.SS.Repository
	if service.Created(&repository) {
		err := client.SsClient.DeleteRepository(repository.ProjectName, repository.Name)
		if err != nil {
			return err
		}
	}

	client.SS.Repository = harborModels.RepositoryPublic{}
	return client.UpdateDB(id, "repository", harborModels.RepositoryPublic{})
}

func (client *Client) DeleteWebhook(id string) error {
	// since webhook is shared, we never actually delete it in harbor

	client.SS.Webhook = harborModels.WebhookPublic{}
	return client.UpdateDB(id, "webhook", harborModels.WebhookPublic{})
}
