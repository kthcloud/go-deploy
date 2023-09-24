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

func TestGetList(t *testing.T) {
	resp := e2e.DoGetRequest(t, "/jobs")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var jobsRead []body.JobRead
	err := e2e.ReadResponseBody(t, resp, &jobsRead)
	assert.NoError(t, err, "jobs were not fetched")

	assert.NotEmpty(t, jobsRead, "jobs were empty")
	for _, job := range jobsRead {
		assert.NotEmpty(t, job.ID, "job id was empty")
		assert.NotEmpty(t, job.Type, "job type was empty")
		assert.NotEmpty(t, job.UserID, "job user id was empty")
	}
}

func TestGet(t *testing.T) {

	// we can't create a job with the api, so we need to trigger a job
	// simplest way is to just create a deployment

	_, jobID := e2e.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName("e2e")})

	resp := e2e.DoGetRequest(t, "/jobs/"+jobID)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var jobRead body.JobRead
	err := e2e.ReadResponseBody(t, resp, &jobRead)
	assert.NoError(t, err, "job was not fetched")

	assert.Equal(t, jobID, jobRead.ID)
	assert.Equal(t, jobRead.Type, job.TypeCreateDeployment)
	assert.Equal(t, jobRead.UserID, e2e.TestUserID)
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
