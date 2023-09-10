package jobs

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	"go-deploy/pkg/conf"
	"go-deploy/test/e2e"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	e2e.Setup()
	code := m.Run()
	e2e.Shutdown()
	os.Exit(code)
}
func TestFetchJobs(t *testing.T) {

	// we can't create a job with the api, so we need to trigger a job
	// simplest way is to just create a deployment

	zone := conf.Env.VM.Zones[0]

	envValue := uuid.NewString()

	requestBody := body.DeploymentCreate{
		Name:    "e2e-" + strings.ReplaceAll(uuid.NewString()[:10], "-", ""),
		Private: false,
		Envs: []body.Env{
			{
				Name:  "e2e",
				Value: envValue,
			},
		},
		GitHub: nil,
		Zone:   &zone.Name,
	}

	resp := e2e.DoPostRequest(t, "/deployments", requestBody)
	if !assert.Equal(t, http.StatusCreated, resp.StatusCode, "deployment was not created") {
		assert.FailNow(t, "deployment was not created")
	}

	var deploymentCreated body.DeploymentCreated
	err := e2e.ReadResponseBody(t, resp, &deploymentCreated)
	assert.NoError(t, err, "deployment was not created")

	t.Cleanup(func() {
		resp = e2e.DoDeleteRequest(t, "/deployments/"+deploymentCreated.ID)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "deployment was not deleted")
		if !assert.Equal(t, http.StatusOK, resp.StatusCode, "deployment was not deleted") {
			assert.FailNow(t, "deployment was not deleted")
		}
	})

	jobID := deploymentCreated.JobID

	resp = e2e.DoGetRequest(t, "/jobs/"+jobID)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var jobRead body.JobRead
	err = e2e.ReadResponseBody(t, resp, &jobRead)
	assert.NoError(t, err, "job was not fetched")

	assert.Equal(t, jobID, jobRead.ID)
	assert.Equal(t, jobRead.Type, "createDeployment")
	assert.Equal(t, jobRead.UserID, e2e.TestUserID)
}
