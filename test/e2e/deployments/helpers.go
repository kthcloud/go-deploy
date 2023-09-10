package deployments

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	"go-deploy/pkg/status_codes"
	"go-deploy/test/e2e"
	"net/http"
	"testing"
	"time"
)

func waitForDeploymentRunning(t *testing.T, id string, callback func(*body.DeploymentRead) bool) {
	loops := 0
	for {
		time.Sleep(10 * time.Second)

		resp := e2e.DoGetRequest(t, "/deployments/"+id)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var deploymentRead body.DeploymentRead
		err := e2e.ReadResponseBody(t, resp, &deploymentRead)
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

		resp := e2e.DoGetRequest(t, "/deployments/"+id)
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

func checkUpDeployment(t *testing.T, url string) bool {
	t.Helper()

	resp, err := http.Get(url)
	if err == nil {
		if resp.StatusCode == http.StatusOK {
			return true
		}
	}

	return false
}

func withDeployment(t *testing.T, requestBody body.DeploymentCreate) body.DeploymentRead {
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

	waitForDeploymentRunning(t, deploymentCreated.ID, func(deploymentRead *body.DeploymentRead) bool {
		//make sure it is accessible
		if deploymentRead.URL != nil {
			return checkUpDeployment(t, *deploymentRead.URL)
		}
		return false
	})

	var deploymentRead body.DeploymentRead
	readResp := e2e.DoGetRequest(t, "/deployments/"+deploymentCreated.ID)
	err = e2e.ReadResponseBody(t, readResp, &deploymentRead)
	assert.NoError(t, err, "deployment was not created")

	assert.NotEmpty(t, deploymentRead.ID)
	assert.Equal(t, requestBody.Name, deploymentRead.Name)
	assert.Equal(t, requestBody.Private, deploymentRead.Private)
	if requestBody.GitHub == nil {
		assert.Empty(t, deploymentRead.Integrations)
	} else {
		assert.NotEmpty(t, deploymentRead.Integrations)
	}

	if requestBody.Zone == nil {
		// some zone is set by default
		assert.NotEmpty(t, deploymentRead.Zone)
	} else {
		assert.Equal(t, requestBody.Zone, deploymentRead.Zone)
	}

	if requestBody.InitCommands == nil {
		assert.Empty(t, deploymentRead.InitCommands)
	} else {
		assert.Equal(t, requestBody.InitCommands, deploymentRead.InitCommands)
	}

	if requestBody.Envs == nil {
		assert.Empty(t, deploymentRead.Envs)
	} else {
		assert.Equal(t, requestBody.Envs, deploymentRead.Envs)
	}

	if requestBody.Volumes == nil {
		assert.Empty(t, deploymentRead.Volumes)
	} else {
		assert.Equal(t, requestBody.Volumes, deploymentRead.Volumes)
	}

	return deploymentRead
}

func withAssumedFailedDeployment(t *testing.T, requestBody body.DeploymentCreate) {
	resp := e2e.DoPostRequest(t, "/deployments", requestBody)
	if resp.StatusCode == http.StatusBadRequest {
		return
	}

	var deploymentCreated body.DeploymentCreated
	err := e2e.ReadResponseBody(t, resp, &deploymentCreated)
	assert.NoError(t, err, "deployment created body was not read")

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

	assert.FailNow(t, "deployment was created but should have failed")
}
