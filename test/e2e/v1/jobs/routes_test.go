package jobs

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v1/body"
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

	_, jobID := v1.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName()})

	job := v1.GetJob(t, jobID)

	assert.Equal(t, jobID, job.ID)
	assert.Equal(t, job.Type, model.JobCreateDeployment)
	assert.Equal(t, job.UserID, e2e.AdminUserID)
}

func TestList(t *testing.T) {
	t.Parallel()

	queries := []string{
		// all
		"?all=true&pageSize=10",
		// by status
		"?status=completed&pageSize=10",
		// by user id
		"?userId=" + e2e.AdminUserID + "&pageSize=10",
	}

	for _, query := range queries {
		jobs := v1.ListJobs(t, query)
		assert.NotEmpty(t, jobs, "jobs were not fetched for query %s. it should have at least one job", query)
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	// We can't create a job with the api, so we need to trigger a job
	// The simplest way is to just create a deployment

	_, jobID := v1.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName()})
	v1.WaitForJobFinished(t, jobID, nil)

	// The job above is assumed to NOT be terminated, so when we update it to terminated, we will notice the change
	terminatedStatus := model.JobStatusTerminated

	v1.UpdateJob(t, jobID, body.JobUpdate{Status: &terminatedStatus})
}
