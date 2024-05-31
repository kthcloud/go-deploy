package jobs

import (
	"github.com/stretchr/testify/assert"
	body2 "go-deploy/dto/v2/body"
	"go-deploy/models/model"
	"go-deploy/test/e2e"
	"go-deploy/test/e2e/v1"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	e2e.Setup()
	code := m.Run()
	e2e.Shutdown()
	os.Exit(code)
}

func TestGet(t *testing.T) {
	t.Parallel()

	// We can't create a job with the API, so we need to trigger a job
	// The simplest way is to create a deployment

	deployment, jobID := v1.WithDeployment(t, body2.DeploymentCreate{Name: e2e.GenName()})

	job := v1.GetJob(t, jobID)

	assert.Equal(t, jobID, job.ID)
	assert.Equal(t, job.Type, model.JobCreateDeployment)
	assert.Equal(t, job.UserID, deployment.OwnerID)
}

func TestList(t *testing.T) {
	t.Parallel()

	queries := []string{
		"?all=true&pageSize=10",
		"?status=completed&pageSize=10",
		"?userId=" + model.TestAdminUserID + "&pageSize=10",
	}

	for _, query := range queries {
		v1.ListJobs(t, query)
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	// Updating a job is only allowed by admins, so we need to use the admin user

	// We can't create a job with the API, so we need to trigger a job
	// The simplest way is to just create a deployment

	_, jobID := v1.WithDeployment(t, body2.DeploymentCreate{Name: e2e.GenName()})
	v1.WaitForJobFinished(t, jobID, nil)

	// The job above is assumed to NOT be terminated, so when we update it to terminated, we will notice the change
	terminatedStatus := model.JobStatusTerminated

	v1.UpdateJob(t, jobID, body2.JobUpdate{Status: &terminatedStatus}, e2e.AdminUser)
}
