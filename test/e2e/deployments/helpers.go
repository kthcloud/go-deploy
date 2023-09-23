package deployments

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	"go-deploy/pkg/status_codes"
	"go-deploy/test/e2e"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func waitForJobFinished(t *testing.T, id string, callback func(*body.JobRead) bool) {
	loops := 0
	for {
		time.Sleep(10 * time.Second)

		resp := e2e.DoGetRequest(t, "/jobs/"+id)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var jobRead body.JobRead
		err := e2e.ReadResponseBody(t, resp, &jobRead)
		if err != nil {
			continue
		}

		if jobRead.Status == status_codes.GetMsg(status_codes.JobFinished) {
			finished := callback(&jobRead)
			if finished {
				break
			}
		}

		if jobRead.Status == status_codes.GetMsg(status_codes.JobTerminated) {
			finished := callback(&jobRead)
			if finished {
				break
			}
		}

		loops++
		if !assert.LessOrEqual(t, loops, 30, "job did not finish in time") {
			assert.FailNow(t, "job did not finish in time")
			break
		}
	}
}

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

func waitForStorageManagerRunning(t *testing.T, id string, callback func(read *body.StorageManagerRead) bool) {
	loops := 0
	for {
		time.Sleep(10 * time.Second)

		resp := e2e.DoGetRequest(t, "/storageManagers/"+id)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var storageManagerRead body.StorageManagerRead
		err := e2e.ReadResponseBody(t, resp, &storageManagerRead)
		if err != nil {
			continue
		}

		finished := callback(&storageManagerRead)
		if finished {
			break
		}

		loops++
		if !assert.LessOrEqual(t, loops, 30, "storage manager did not start in time") {
			assert.FailNow(t, "storage manager did not start in time")
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

func checkUpURL(t *testing.T, url string) bool {
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
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			assert.FailNow(t, "resource was not deleted")
		}

		waitForDeploymentDeleted(t, deploymentCreated.ID, func() bool {
			return true
		})
	})

	waitForJobFinished(t, deploymentCreated.JobID, func(jobRead *body.JobRead) bool {
		return true
	})

	waitForDeploymentRunning(t, deploymentCreated.ID, func(deploymentRead *body.DeploymentRead) bool {
		//make sure it is accessible
		if deploymentRead.URL != nil {
			return checkUpURL(t, *deploymentRead.URL)
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

	// PORT env is generated, so we check for that first, then delete and match exact with the rest
	assert.NotEmpty(t, deploymentRead.Envs)

	portRequested := 0
	customPort := false
	for _, env := range requestBody.Envs {
		if env.Name == "PORT" {
			portRequested, _ = strconv.Atoi(env.Value)
			customPort = true
			break
		}
	}

	if portRequested == 0 {
		portRequested = 8080
	}

	portReceived := 0
	for _, env := range deploymentRead.Envs {
		if env.Name == "PORT" {
			portReceived, _ = strconv.Atoi(env.Value)
			break
		}
	}

	assert.NotZero(t, portReceived)
	assert.Equal(t, portRequested, portReceived)

	if !customPort {
		// remove PORT from read, since we won't be able to match the rest of the envs other wise
		for i, env := range deploymentRead.Envs {
			if env.Name == "PORT" {
				deploymentRead.Envs = append(deploymentRead.Envs[:i], deploymentRead.Envs[i+1:]...)
				break
			}
		}
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
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			assert.FailNow(t, "resource was not deleted")
		}

		waitForDeploymentDeleted(t, deploymentCreated.ID, func() bool {
			return true
		})
	})

	assert.FailNow(t, "deployment was created but should have failed")
}
