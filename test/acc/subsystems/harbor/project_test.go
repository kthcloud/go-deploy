package harbor

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/pkg/subsystems/harbor/models"
	"testing"
)

func TestCreateProject(t *testing.T) {
	project := withHarborProject(t)
	cleanUpHarborProject(t, project.ID)
}

func TestProjectWithBadWebhook(t *testing.T) {

	client := withHarborClient(t)
	project := withHarborProject(t)

	public := &models.WebhookPublic{
		Name:        project.Name + "-webhook",
		ProjectID:   project.ID,
		ProjectName: project.Name,
		Target:      "not http",
		Token:       "some token",
	}

	_, err := client.CreateWebhook(public)
	assert.Error(t, err, "expected error when creating webhook with bad target")

	cleanUpHarborProject(t, project.ID)
}

func TestUpdateProject(t *testing.T) {
	client := withHarborClient(t)
	project := withHarborProject(t)

	err := client.UpdateProject(project)
	assert.NoError(t, err, "failed to update harbor project")

	updatedProject, err := client.ReadProject(project.ID)
	assert.NoError(t, err, "failed to read harbor project")
	assert.EqualValues(t, project, updatedProject, "updated project does not match")

	cleanUpHarborProject(t, project.ID)
}
