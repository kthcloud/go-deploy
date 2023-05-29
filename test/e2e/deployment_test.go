package e2e

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	"go-deploy/pkg/status_codes"
	"net/http"
	"strings"
	"testing"
	"time"
)

func waitForDeploymentCreated(t *testing.T, id string, callback func(*body.DeploymentRead) bool) {
	loops := 0
	for {
		time.Sleep(10 * time.Second)

		resp := doGetRequest(t, "/deployments/"+id)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var deploymentRead body.DeploymentRead
		err := readResponseBody(t, resp, &deploymentRead)
		if err != nil {
			continue
		}

		if deploymentRead.Status == status_codes.GetMsg(status_codes.ResourceRunning) {
			finished := callback(&deploymentRead)
			if finished {
				break
			}
		}

		loops++
		if !assert.LessOrEqual(t, loops, 30, "deployment did not start in time") {
			assert.FailNow(t, "deployment did not start in time")
			break
		}
	}
}

func waitForDeploymentDeleted(t *testing.T, id string, callback func() bool) {
	loops := 0
	for {
		time.Sleep(10 * time.Second)

		resp := doGetRequest(t, "/deployments/"+id)
		if resp.StatusCode == http.StatusNotFound {
			if callback() {
				break
			}
			break
		}

		loops++
		if !assert.LessOrEqual(t, loops, 30, "deployment did not start in time") {
			assert.FailNow(t, "deployment did not start in time")
			break
		}
	}
}

func TestGetDeployments(t *testing.T) {
	setup(t)
	withServer(t)

	resp := doGetRequest(t, "/deployments")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCreateDeployment(t *testing.T) {
	setup(t)
	withServer(t)

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
	}

	resp := doPostRequest(t, "/deployments", requestBody)
	if !assert.Equal(t, http.StatusCreated, resp.StatusCode, "deployment was not created") {
		assert.FailNow(t, "deployment was not created")
	}

	var deploymentCreated body.DeploymentCreated
	err := readResponseBody(t, resp, &deploymentCreated)
	assert.NoError(t, err, "deployment was not created")

	t.Cleanup(func() {
		resp = doDeleteRequest(t, "/deployments/"+deploymentCreated.ID)
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
