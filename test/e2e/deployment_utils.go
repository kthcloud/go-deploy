package e2e

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	status_codes2 "go-deploy/pkg/app/status_codes"
	"golang.org/x/net/idna"
	"net/http"
	"strconv"
	"testing"
)

func WaitForDeploymentRunning(t *testing.T, id string, callback func(*body.DeploymentRead) bool) {
	fetchUntil(t, "/deployments/"+id, func(resp *http.Response) bool {
		var deploymentRead body.DeploymentRead
		err := ReadResponseBody(t, resp, &deploymentRead)
		assert.NoError(t, err, "deployment was not fetched")

		if deploymentRead.Status == status_codes2.GetMsg(status_codes2.ResourceRunning) {
			if callback == nil || callback(&deploymentRead) {
				return true
			}
		}

		return false
	})
}

func WaitForSmRunning(t *testing.T, id string, callback func(read *body.SmRead) bool) {
	fetchUntil(t, "/storageManagers/"+id, func(resp *http.Response) bool {
		var smRead body.SmRead
		err := ReadResponseBody(t, resp, &smRead)
		assert.NoError(t, err, "storage manager was not fetched")

		if callback == nil {
			return true
		}

		return callback(&smRead)
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

func GetDeployment(t *testing.T, id string, userID ...string) body.DeploymentRead {
	resp := DoGetRequest(t, "/deployments/"+id, userID...)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "deployment was not fetched")

	var deploymentRead body.DeploymentRead
	err := ReadResponseBody(t, resp, &deploymentRead)
	assert.NoError(t, err, "deployment was not fetched")

	return deploymentRead
}

func ListDeployments(t *testing.T, query string, userID ...string) []body.DeploymentRead {
	resp := DoGetRequest(t, "/deployments"+query, userID...)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "deployments were not fetched")

	var deployments []body.DeploymentRead
	err := ReadResponseBody(t, resp, &deployments)
	assert.NoError(t, err, "deployments were not fetched")

	return deployments
}

func WithDeployment(t *testing.T, requestBody body.DeploymentCreate) (body.DeploymentRead, string) {
	resp := DoPostRequest(t, "/deployments", requestBody)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "deployment was not created")

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
		// Some zone is set by default
		assert.NotEmpty(t, deploymentRead.Zone)
	} else {
		assert.Equal(t, requestBody.Zone, deploymentRead.Zone)
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

	EqualOrEmpty(t, requestBody.InitCommands, deploymentRead.InitCommands, "invalid init commands")
	EqualOrEmpty(t, requestBody.Envs, deploymentRead.Envs, "invalid envs")
	EqualOrEmpty(t, requestBody.Volumes, deploymentRead.Volumes, "invalid volumes")

	if requestBody.CustomDomain != nil {
		punyEncoded, err := idna.New().ToASCII("https://" + *requestBody.CustomDomain)
		assert.NoError(t, err, "custom domain was not puny encoded")
		assert.Equal(t, punyEncoded, *deploymentRead.CustomDomainURL)
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
