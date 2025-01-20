package jobs

import (
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/test/e2e"
	"github.com/kthcloud/go-deploy/test/e2e/v2"
	"github.com/stretchr/testify/assert"
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

	deployment, jobID := v2.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName()})

	job := v2.GetJob(t, jobID)

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
		v2.ListJobs(t, query)
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	// Updating a job is only allowed by admins, so we need to use the admin user

	// We can't create a job with the API, so we need to trigger a job
	// The simplest way is to just create a deployment

	_, jobID := v2.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName()})
	v2.WaitForJobFinished(t, jobID, nil)

	// The job above is assumed to NOT be terminated, so when we update it to terminated, we will notice the change
	terminatedStatus := model.JobStatusTerminated

	v2.UpdateJob(t, jobID, body.JobUpdate{Status: &terminatedStatus}, e2e.AdminUser)
}
