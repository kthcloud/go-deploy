package v2

import (
	"crypto/tls"
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v2/body"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/test"
	"go-deploy/test/e2e"
	"golang.org/x/net/idna"
	"net/http"
	"strconv"
	"testing"
)

const (
	DeploymentPath  = "/v2/deployments/"
	DeploymentsPath = "/v2/deployments"
)

func CheckUpURL(t *testing.T, url string) bool {
	t.Helper()

	tr := &http.Transport{
		// Local environments does not have valid certificates
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	resp, err := client.Get(url)
	if err == nil {
		if resp.StatusCode == http.StatusOK {
			return true
		}
	}

	return false
}

func GetDeployment(t *testing.T, id string, user ...string) body.DeploymentRead {
	resp := e2e.DoGetRequest(t, DeploymentPath+id, user...)
	return e2e.MustParse[body.DeploymentRead](t, resp)
}

func ListDeployments(t *testing.T, query string, user ...string) []body.DeploymentRead {
	resp := e2e.DoGetRequest(t, DeploymentsPath+query, user...)
	return e2e.MustParse[[]body.DeploymentRead](t, resp)
}

func UpdateDeployment(t *testing.T, id string, requestBody body.DeploymentUpdate, user ...string) body.DeploymentRead {
	resp := e2e.DoPostRequest(t, DeploymentPath+id, requestBody, user...)
	deploymentUpdated := e2e.MustParse[body.DeploymentUpdated](t, resp)

	if deploymentUpdated.JobID != nil {
		WaitForJobFinished(t, *deploymentUpdated.JobID, nil)
	}
	WaitForDeploymentRunning(t, id, func(read *body.DeploymentRead) bool {
		if read.Private {
			return true
		}

		// Make sure it is accessible
		if read.URL != nil {
			return CheckUpURL(t, *read.URL)
		}
		return false
	})

	updated := GetDeployment(t, id, user...)

	if requestBody.CustomDomain != nil {
		punyEncoded, err := idna.New().ToASCII("https://" + *requestBody.CustomDomain)
		assert.NoError(t, err, "custom domain was not puny encoded")
		assert.Equal(t, punyEncoded, updated.CustomDomain.URL)
	}

	if requestBody.Image != nil {
		assert.Equal(t, *requestBody.Image, *updated.Image)
	}

	if requestBody.InitCommands != nil {
		test.EqualOrEmpty(t, *requestBody.InitCommands, updated.InitCommands)
	}

	if requestBody.Envs != nil {
		// PORT env is generated, so it will always be in the read, but not necessarily in the request
		customPort := false
		for _, env := range *requestBody.Envs {
			if env.Name == "PORT" {
				customPort = true
				break
			}
		}

		// If it was not requested, we add it to the request body so we can compare the rest
		if !customPort {
			*requestBody.Envs = append(*requestBody.Envs, body.Env{
				Name:  "PORT",
				Value: "8080",
			})
		}

		test.EqualOrEmpty(t, *requestBody.Envs, updated.Envs)
	}

	if requestBody.Volumes != nil {
		test.EqualOrEmpty(t, *requestBody.Volumes, updated.Volumes)
	}

	if requestBody.Private != nil {
		assert.Equal(t, *requestBody.Private, updated.Private)
	}

	if requestBody.HealthCheckPath != nil {
		assert.Equal(t, *requestBody.HealthCheckPath, updated.HealthCheckPath)
	}

	if requestBody.Replicas != nil {
		assert.Equal(t, *requestBody.Replicas, updated.Specs.Replicas)
	}

	return updated
}

func WithDeployment(t *testing.T, requestBody body.DeploymentCreate, user ...string) (body.DeploymentRead, string) {
	resp := e2e.DoPostRequest(t, DeploymentsPath, requestBody, user...)
	deploymentCreated := e2e.MustParse[body.DeploymentCreated](t, resp)
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

	deploymentRead := GetDeployment(t, deploymentCreated.ID)
	assert.NotEmpty(t, deploymentRead.ID)
	assert.Equal(t, requestBody.Name, deploymentRead.Name)
	assert.Equal(t, requestBody.Private, deploymentRead.Private)

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

	test.EqualOrEmpty(t, requestBody.InitCommands, deploymentRead.InitCommands, "invalid init commands")
	test.EqualOrEmpty(t, requestBody.Args, deploymentRead.Args, "invalid args")
	test.EqualOrEmpty(t, requestBody.Envs, deploymentRead.Envs, "invalid envs")
	test.EqualOrEmpty(t, requestBody.Volumes, deploymentRead.Volumes, "invalid volumes")

	if requestBody.CustomDomain != nil {
		punyEncoded, err := idna.New().ToASCII("https://" + *requestBody.CustomDomain)
		assert.NoError(t, err, "custom domain was not puny encoded")
		assert.Contains(t, deploymentRead.CustomDomain.URL, punyEncoded, "custom domain was not set")
	}

	return deploymentRead, deploymentCreated.JobID
}

func WithAssumedFailedDeployment(t *testing.T, requestBody body.DeploymentCreate, user ...string) {
	resp := e2e.DoPostRequest(t, DeploymentsPath, requestBody, user...)
	if e2e.IsUserError(resp.StatusCode) {
		return
	}

	var deploymentCreated body.DeploymentCreated
	err := e2e.ReadResponseBody(t, resp, &deploymentCreated)
	assert.NoError(t, err, "deployment created body was not read")

	t.Cleanup(func() { CleanUpDeployment(t, deploymentCreated.ID) })

	assert.FailNow(t, "deployment was created but should have failed")
}

func CleanUpDeployment(t *testing.T, id string) {
	resp := e2e.DoDeleteRequest(t, DeploymentPath+id, e2e.AdminUser)
	if resp.StatusCode == http.StatusNotFound {
		return
	}

	if resp.StatusCode == http.StatusOK {
		deploymentDeleted := e2e.MustParse[body.DeploymentDeleted](t, resp)
		WaitForJobFinished(t, deploymentDeleted.JobID, nil)
		WaitForDeploymentDeleted(t, deploymentDeleted.ID, nil)

		return
	}

	assert.FailNow(t, "deployment was not deleted")
}

func WaitForDeploymentRunning(t *testing.T, id string, callback func(*body.DeploymentRead) bool) {
	e2e.FetchUntil(t, DeploymentPath+id, func(resp *http.Response) bool {
		deploymentRead := e2e.MustParse[body.DeploymentRead](t, resp)
		if deploymentRead.Status == status_codes.GetMsg(status_codes.ResourceRunning) {
			if callback == nil || callback(&deploymentRead) {
				return true
			}
		}

		return false
	})
}

func WaitForDeploymentDeleted(t *testing.T, id string, callback func() bool) {
	e2e.FetchUntil(t, DeploymentPath+id, func(resp *http.Response) bool {
		return resp.StatusCode == http.StatusNotFound
	})
}
