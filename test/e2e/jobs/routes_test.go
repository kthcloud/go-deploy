package jobs

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	"go-deploy/models/sys/job"
	"go-deploy/test/e2e"
	"net/http"
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

	// We can't create a job with the API, so we need to trigger a job
	// The simplest way is to create a deployment

	_, jobID := e2e.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName("e2e")})

	resp := e2e.DoGetRequest(t, "/jobs/"+jobID)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var jobRead body.JobRead
	err := e2e.ReadResponseBody(t, resp, &jobRead)
	assert.NoError(t, err, "job was not fetched")

	assert.Equal(t, jobID, jobRead.ID)
	assert.Equal(t, jobRead.Type, job.TypeCreateDeployment)
	assert.Equal(t, jobRead.UserID, e2e.AdminUserID)
}

func TestList(t *testing.T) {
	queries := []string{
		// all
		"?all=true&pageSize=10",
		// by status
		"?status=finished&pageSize=10",
		// by user id
		"?userID=" + e2e.AdminUserID + "&pageSize=10",
	}

	for _, query := range queries {
		jobs := e2e.ListJobs(t, query)
		assert.NotEmpty(t, jobs, "jobs were not fetched. it should have at least one job")
	}
}

func TestUpdate(t *testing.T) {
	// we can't create a job with the api, so we need to trigger a job
	// simplest way is to just create a deployment

	_, jobID := e2e.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName("e2e")})
	e2e.WaitForJobFinished(t, jobID, nil)

	terminatedStatus := job.StatusTerminated
	resp := e2e.DoPostRequest(t, "/jobs/"+jobID, body.JobUpdate{Status: &terminatedStatus})
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var jobRead body.JobRead
	err := e2e.ReadResponseBody(t, resp, &jobRead)
	assert.NoError(t, err, "job was not updated")

	assert.Equal(t, job.StatusTerminated, jobRead.Status)
}
