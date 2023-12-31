package harbor

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/harbor"
	"go-deploy/pkg/subsystems/harbor/models"
	"go-deploy/test"
	"go-deploy/test/acc"
	"strings"
	"testing"
)

func withContext(t *testing.T) (*harbor.Client, *models.ProjectPublic) {
	p := withDefaultProject(t)
	return withClient(t, p.Name), p
}

func withClient(t *testing.T, projectName string) *harbor.Client {
	client, err := harbor.New(&harbor.ClientConf{
		URL:      config.Config.Harbor.URL,
		Username: config.Config.Harbor.User,
		Password: config.Config.Harbor.Password,
		Project:  projectName,
	})

	test.NoError(t, err, "failed to create harbor client")
	assert.NotNil(t, client, "harbor client is nil")

	return client
}

func withDefaultProject(t *testing.T) *models.ProjectPublic {
	p := &models.ProjectPublic{
		Name: acc.GenName(),
	}

	return withProject(t, p)
}

func withProject(t *testing.T, p *models.ProjectPublic) *models.ProjectPublic {
	c := withClient(t, "")

	pCreated, err := c.CreateProject(p)
	test.NoError(t, err, "failed to create harbor project")
	t.Cleanup(func() { cleanUpProject(t, c, pCreated.ID) })

	assert.NotEmpty(t, pCreated.ID, "no id received from client")
	assert.Equal(t, p.Name, pCreated.Name, "project name does not match")
	assert.Equal(t, p.Public, pCreated.Public, "project public does not match")

	c.Project = pCreated.Name

	return pCreated
}

func withDefaultRobot(t *testing.T, c *harbor.Client) *models.RobotPublic {
	r := &models.RobotPublic{
		Name: acc.GenName(),
	}

	return withRobot(t, c, r)
}

func withRobot(t *testing.T, c *harbor.Client, r *models.RobotPublic) *models.RobotPublic {
	rCreated, err := c.CreateRobot(r)
	test.NoError(t, err, "failed to create harbor robot")
	t.Cleanup(func() { cleanUpRobot(t, c, rCreated.ID) })

	assert.NotEmpty(t, rCreated.ID, "no id received from client")
	assert.Equal(t, r.Name, rCreated.Name, "robot name does not match")

	return rCreated
}

func withDefaultWebhook(t *testing.T, c *harbor.Client) *models.WebhookPublic {
	w := &models.WebhookPublic{
		Name:   acc.GenName(),
		Target: "https://some-url.com",
		Token:  acc.GenName(),
	}

	return withWebhook(t, c, w)
}

func withWebhook(t *testing.T, c *harbor.Client, w *models.WebhookPublic) *models.WebhookPublic {
	wCreated, err := c.CreateWebhook(w)
	test.NoError(t, err, "failed to create harbor webhook")
	t.Cleanup(func() { cleanUpWebhook(t, c, wCreated.ID) })

	assert.NotEmpty(t, wCreated.ID, "no id received from client")
	assert.Equal(t, w.Name, wCreated.Name, "webhook name does not match")
	assert.Equal(t, w.Target, wCreated.Target, "webhook target does not match")
	assert.Equal(t, w.Token, wCreated.Token, "webhook token does not match")

	return wCreated
}

func withDefaultRepository(t *testing.T, c *harbor.Client) *models.RepositoryPublic {
	splits := strings.Split(config.Config.Registry.PlaceholderImage, "/")
	project := splits[len(splits)-2]
	repository := splits[len(splits)-1]

	r := &models.RepositoryPublic{
		Name: acc.GenName(),
		Placeholder: &models.PlaceHolder{
			ProjectName:    project,
			RepositoryName: repository,
		},
	}

	return withRepository(t, c, r)
}

func withRepository(t *testing.T, c *harbor.Client, r *models.RepositoryPublic) *models.RepositoryPublic {
	rCreated, err := c.CreateRepository(r)
	test.NoError(t, err, "failed to create harbor repository")
	t.Cleanup(func() { cleanUpRepository(t, c, rCreated.Name) })

	assert.NotEmpty(t, rCreated.ID, "no id received from client")
	assert.Equal(t, r.Name, rCreated.Name, "repository name does not match")
	assert.True(t, rCreated.Seeded, "repository is not seeded")

	return rCreated
}

func cleanUpProject(t *testing.T, c *harbor.Client, id int) {
	err := c.DeleteProject(id)
	test.NoError(t, err, "failed to delete harbor project")

	deletedProject, err := c.ReadProject(id)
	test.NoError(t, err, "failed to read harbor project")
	assert.Nil(t, deletedProject, "failed to delete harbor project")

	err = c.DeleteProject(id)
	test.NoError(t, err, "failed to delete harbor project")
}

func cleanUpRobot(t *testing.T, c *harbor.Client, id int) {
	err := c.DeleteRobot(id)
	test.NoError(t, err, "failed to delete harbor robot")

	deletedRobot, err := c.ReadRobot(id)
	test.NoError(t, err, "failed to read harbor robot")
	assert.Nil(t, deletedRobot, "failed to delete harbor robot")

	err = c.DeleteRobot(id)
	test.NoError(t, err, "failed to delete harbor robot")
}

func cleanUpWebhook(t *testing.T, c *harbor.Client, id int) {
	err := c.DeleteWebhook(id)
	test.NoError(t, err, "failed to delete harbor webhook")

	deletedWebhook, err := c.ReadWebhook(id)
	test.NoError(t, err, "failed to read harbor webhook")
	assert.Nil(t, deletedWebhook, "failed to delete harbor webhook")

	err = c.DeleteWebhook(id)
	test.NoError(t, err, "failed to delete harbor webhook")
}

func cleanUpRepository(t *testing.T, c *harbor.Client, name string) {
	err := c.DeleteRepository(name)
	test.NoError(t, err, "failed to delete harbor repository")

	deletedRepository, err := c.ReadRepository(name)
	test.NoError(t, err, "failed to read harbor repository")
	assert.Nil(t, deletedRepository, "failed to delete harbor repository")

	err = c.DeleteRepository(name)
	test.NoError(t, err, "failed to delete harbor repository")
}
