package harbor

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/harbor"
	"go-deploy/pkg/subsystems/harbor/models"
	"testing"
)

func withHarborClient(t *testing.T) *harbor.Client {
	client, err := harbor.New(&harbor.ClientConf{
		ApiUrl:   conf.Env.Harbor.URL,
		Username: conf.Env.Harbor.User,
		Password: conf.Env.Harbor.Password,
	})

	assert.NoError(t, err, "failed to create harbor client")
	assert.NotNil(t, client, "harbor client is nil")

	return client
}

func withHarborProject(t *testing.T) *models.ProjectPublic {
	client := withHarborClient(t)

	project := &models.ProjectPublic{
		Name: "acc-test-" + uuid.New().String(),
	}

	id, err := client.CreateProject(project)
	assert.NoError(t, err, "failed to create harbor project")

	createdProject, err := client.ReadProject(id)
	assert.NoError(t, err, "failed to read harbor project")
	assert.NotNil(t, createdProject, "harbor project is nil")

	project.ID = id
	assert.NotZero(t, project.ID, "no id received from client")

	assert.EqualValues(t, project, createdProject, "created project does not match")

	return project
}

func withHarborRobot(t *testing.T, project *models.ProjectPublic) *models.RobotPublic {
	client := withHarborClient(t)

	robot := &models.RobotPublic{
		Name:        "acc-test-" + uuid.New().String(),
		ProjectID:   project.ID,
		ProjectName: project.Name,
		Disable:     false,
		Secret:      "some secret",
	}

	id, err := client.CreateRobot(robot)
	assert.NoError(t, err, "failed to create harbor robot")

	createdRobot, err := client.ReadRobot(id)
	assert.NoError(t, err, "failed to read harbor robot")

	robot.ID = id
	assert.NotZero(t, robot.ID, "no id received from client")

	robot.HarborName = createdRobot.HarborName
	assert.NotEmpty(t, robot.HarborName, "no harbor name received from client")

	// we can't verify the secret
	secret := robot.Secret
	robot.Secret = ""

	createdSecret := createdRobot.Secret
	createdRobot.Secret = ""

	assert.EqualValues(t, robot, createdRobot, "created robot does not match")

	robot.Secret = secret
	createdRobot.Secret = createdSecret

	return robot
}

func withHarborWebhook(t *testing.T, project *models.ProjectPublic) *models.WebhookPublic {
	client := withHarborClient(t)

	webhook := &models.WebhookPublic{
		Name:        "acc-test-" + uuid.New().String(),
		ProjectID:   project.ID,
		ProjectName: project.Name,
		Target:      "https://some-url.com",
		Token:       "acc-test-" + uuid.New().String(),
	}

	id, err := client.CreateWebhook(webhook)
	assert.NoError(t, err, "failed to create harbor webhook")
	assert.NotZero(t, id, "no id received from client")

	createdWebhook, err := client.ReadWebhook(project.ID, id)
	assert.NoError(t, err, "failed to read harbor webhook")
	assert.NotNil(t, createdWebhook, "failed to read harbor webhook")

	webhook.ID = createdWebhook.ID
	assert.NotZero(t, webhook.ID, "no id received from client")
	assert.Equal(t, webhook.ID, id, "id does not match")

	assert.EqualValues(t, webhook, createdWebhook, "created webhook does not match")

	return webhook
}

func cleanUpHarborProject(t *testing.T, id int) {
	client := withHarborClient(t)

	err := client.DeleteProject(id)
	assert.NoError(t, err, "failed to delete harbor project")

	deletedProject, err := client.ReadProject(id)
	assert.NoError(t, err, "failed to read harbor project")
	assert.Nil(t, deletedProject, "failed to delete harbor project")

	// should not return error if project is already deleted
	err = client.DeleteProject(id)
	assert.NoError(t, err, "failed to delete harbor project")
}

func cleanUpHarborRobot(t *testing.T, id int) {
	client := withHarborClient(t)

	err := client.DeleteRobot(id)
	assert.NoError(t, err, "failed to delete harbor robot")

	deletedRobot, err := client.ReadRobot(id)
	assert.NoError(t, err, "failed to read harbor robot")
	assert.Nil(t, deletedRobot, "failed to delete harbor robot")

	// should not return error if robot is already deleted
	err = client.DeleteRobot(id)
	assert.NoError(t, err, "failed to delete harbor robot")
}

func cleanUpHarborWebhook(t *testing.T, id int, projectID int) {
	client := withHarborClient(t)

	err := client.DeleteWebhook(id, projectID)
	assert.NoError(t, err, "failed to delete harbor webhook")

	deletedWebhook, err := client.ReadWebhook(id, projectID)
	assert.NoError(t, err, "failed to read harbor webhook")
	assert.Nil(t, deletedWebhook, "failed to delete harbor webhook")

	// should not return error if webhook is already deleted
	err = client.DeleteWebhook(id, projectID)
	assert.NoError(t, err, "failed to delete harbor webhook")
}
