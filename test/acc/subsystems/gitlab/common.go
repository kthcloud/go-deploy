package gitlab

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/gitlab"
	"go-deploy/pkg/subsystems/gitlab/models"
	"go-deploy/test"
	"go-deploy/test/acc"
	"testing"
)

func withClient(t *testing.T) *gitlab.Client {
	client, err := gitlab.New(&gitlab.ClientConf{
		URL:   config.Config.GitLab.URL,
		Token: config.Config.GitLab.Token,
	})

	test.NoError(t, err, "failed to create gitlab client")

	return client
}

func withDefaultProject(t *testing.T) *models.ProjectPublic {
	p := &models.ProjectPublic{
		Name:      acc.GenName(),
		ImportURL: "",
	}

	return withProject(t, p)
}

func withProject(t *testing.T, p *models.ProjectPublic) *models.ProjectPublic {
	client := withClient(t)

	pCreated, err := client.CreateProject(p)
	test.NoError(t, err, "failed to create project")

	assert.Equal(t, p.Name, pCreated.Name, "project name mismatch")
	assert.Equal(t, p.ImportURL, pCreated.ImportURL, "project import url mismatch")

	t.Cleanup(func() { cleanUpProject(t, pCreated.ID) })
	return pCreated
}

func cleanUpProject(t *testing.T, id int) {
	c := withClient(t)

	err := c.DeleteProject(id)
	test.NoError(t, err, "failed to delete project")

	deletedJob, err := c.ReadProject(id)
	test.NoError(t, err, "failed to read project")
	assert.Nil(t, deletedJob, "project not deleted")

	err = c.DeleteProject(id)
	test.NoError(t, err, "failed to delete project")
}
