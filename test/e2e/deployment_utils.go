package e2e

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	"go-deploy/pkg/status_codes"
	"net/http"
	"strconv"
	"testing"
)

func WaitForDeploymentRunning(t *testing.T, id string, callback func(*body.DeploymentRead) bool) {
	fetchUntil(t, "/deployments/"+id, func(resp *http.Response) bool {
		var deploymentRead body.DeploymentRead
		err := ReadResponseBody(t, resp, &deploymentRead)
		assert.NoError(t, err, "deployment was not fetched")

		if deploymentRead.Status == status_codes.GetMsg(status_codes.ResourceRunning) {
			if callback == nil || callback(&deploymentRead) {
				return true
			}
		}

		return false
	})
}

func WaitForStorageManagerRunning(t *testing.T, id string, callback func(read *body.StorageManagerRead) bool) {
	fetchUntil(t, "/storageManagers/"+id, func(resp *http.Response) bool {
		var storageManagerRead body.StorageManagerRead
		err := ReadResponseBody(t, resp, &storageManagerRead)
		assert.NoError(t, err, "storage manager was not fetched")

		if callback == nil {
			return true
		}

		return callback(&storageManagerRead)
	})
}

func WaitForDeploymentDeleted(t *testing.T, id string, callback func() bool) {
	fetchUntil(t, "/deployments/"+id, func(resp *http.Response) bool {
		return resp.StatusCode == http.StatusNotFound
	})
}

func CheckUpURL(t *testing.T, url string) bool {
	t.Helper()

	resp, err := http.Get(url)
	if err == nil {
		if resp.StatusCode == http.StatusOK {
			return true
		}
	}

	return false
}

func WithDeployment(t *testing.T, requestBody body.DeploymentCreate) (body.DeploymentRead, string) {
	resp := DoPostRequest(t, "/deployments", requestBody)
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "deployment was not created")

	var deploymentCreated body.DeploymentCreated
	err := ReadResponseBody(t, resp, &deploymentCreated)
	assert.NoError(t, err, "deployment was not created")

	t.Cleanup(func() {
		CleanUpDeployment(t, deploymentCreated.ID)
	})

	WaitForJobFinished(t, deploymentCreated.JobID, nil)
	WaitForDeploymentRunning(t, deploymentCreated.ID, func(deploymentRead *body.DeploymentRead) bool {
		//make sure it is accessible
		if deploymentRead.URL != nil {
			return CheckUpURL(t, *deploymentRead.URL)
		}
		return false
	})

	var deploymentRead body.DeploymentRead
	readResp := DoGetRequest(t, "/deployments/"+deploymentCreated.ID)
	err = ReadResponseBody(t, readResp, &deploymentRead)
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

	return deploymentRead, deploymentCreated.JobID
}

func WithAssumedFailedDeployment(t *testing.T, requestBody body.DeploymentCreate) {
	resp := DoPostRequest(t, "/deployments", requestBody)
	if resp.StatusCode == http.StatusBadRequest {
		return
	}

	var deploymentCreated body.DeploymentCreated
	err := ReadResponseBody(t, resp, &deploymentCreated)
	assert.NoError(t, err, "deployment created body was not read")

	t.Cleanup(func() { CleanUpDeployment(t, deploymentCreated.ID) })

	assert.FailNow(t, "deployment was created but should have failed")
}

func CleanUpDeployment(t *testing.T, id string) {
	resp := DoDeleteRequest(t, "/deployments/"+id)
	if resp.StatusCode == http.StatusNotFound {
		return
	}

	if resp.StatusCode == http.StatusOK {
		var vmDeleted body.VmDeleted
		err := ReadResponseBody(t, resp, &vmDeleted)
		assert.NoError(t, err, "deleted body was not read")
		assert.Equal(t, id, vmDeleted.ID)

		WaitForJobFinished(t, vmDeleted.JobID, nil)
		WaitForDeploymentDeleted(t, vmDeleted.ID, nil)

		return
	}

	assert.FailNow(t, "deployment was not deleted")
}
