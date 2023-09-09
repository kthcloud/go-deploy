package deployments

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	"go-deploy/pkg/conf"
	"go-deploy/test/e2e"
	"net/http"
	"strings"
	"testing"
)

func TestGetDeployments(t *testing.T) {
	resp := e2e.DoGetRequest(t, "/deployments")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCreateDeployment(t *testing.T) {

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

		waitForDeploymentDeleted(t, deploymentCreated.ID, func() bool {
			return true
		})
	})

	waitForDeploymentCreated(t, deploymentCreated.ID, func(deploymentRead *body.DeploymentRead) bool {
		//make sure it is accessible
		if deploymentRead.URL != nil {
			return checkUpDeployment(t, *deploymentRead.URL)
		}
		return false
	})
}
