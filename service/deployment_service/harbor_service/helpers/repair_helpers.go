package helpers

import (
	harborModels "go-deploy/pkg/subsystems/harbor/models"
	"go-deploy/service"
	"log"
)

func (client *Client) RepairProject(id string, genPublic func() *harborModels.ProjectPublic) error {
	dbProject := client.SS.Project
	if service.NotCreated(&dbProject) {
		public := genPublic()
		if public == nil {
			log.Println("no public supplied for harbor project when trying to create it in the repair process")
			return nil
		}

		_, err := client.CreateProject(id, public)
		return err
	}

	return service.UpdateIfDiff[harborModels.ProjectPublic](
		dbProject,
		func() (*harborModels.ProjectPublic, error) {
			return client.SsClient.ReadProject(dbProject.ID)
		},
		client.SsClient.UpdateProject,
		func(dbResource *harborModels.ProjectPublic) error {
			return client.RecreateProject(id, dbResource)
		},
	)
}

func (client *Client) RepairRepository(id string, genPublic func() *harborModels.RepositoryPublic) error {
	dbRepository := client.SS.Repository
	if service.NotCreated(&dbRepository) {
		public := genPublic()
		if public == nil {
			log.Println("no public supplied for harbor repository when trying to create it in the repair process")
			return nil
		}

		_, err := client.CreateRepository(id, public)
		return err
	}

	return service.UpdateIfDiff[harborModels.RepositoryPublic](
		dbRepository,
		func() (*harborModels.RepositoryPublic, error) {
			return client.SsClient.ReadRepository(dbRepository.ProjectName, dbRepository.Name)
		},
		client.SsClient.UpdateRepository,
		func(dbResource *harborModels.RepositoryPublic) error {
			// don't recreate repository since it would mean a new push to the registry by the user
			return nil
		},
	)
}

func (client *Client) RepairRobot(id string, genPublic func() *harborModels.RobotPublic) error {
	dbRobot := client.SS.Robot
	if service.NotCreated(&dbRobot) {
		public := genPublic()
		if public == nil {
			log.Println("no public supplied for harbor robot when trying to create it in the repair process")
			return nil
		}

		_, err := client.CreateRobot(id, public)
		return err
	}

	return service.UpdateIfDiff[harborModels.RobotPublic](
		dbRobot,
		func() (*harborModels.RobotPublic, error) {
			return client.SsClient.ReadRobot(dbRobot.ID)
		},
		client.SsClient.UpdateRobot,
		func(dbResource *harborModels.RobotPublic) error {
			// don't recreate the robot since the credentials mights be used somewhere
			return nil
		},
	)
}

func (client *Client) RepairWebhook(id string, genPublic func() *harborModels.WebhookPublic) error {
	dbWebhook := client.SS.Webhook
	if service.NotCreated(&dbWebhook) {
		public := genPublic()
		if public == nil {
			log.Println("no public supplied for harbor webhook when trying to create it in the repair process")
			return nil
		}

		_, err := client.CreateWebhook(id, public)
		return err
	}

	return service.UpdateIfDiff[harborModels.WebhookPublic](
		dbWebhook,
		func() (*harborModels.WebhookPublic, error) {
			return client.SsClient.ReadWebhook(dbWebhook.ProjectID, dbWebhook.ID)
		},
		client.SsClient.UpdateWebhook,
		func(dbResource *harborModels.WebhookPublic) error {
			return client.RecreateWebhook(id, dbResource)
		},
	)
}
