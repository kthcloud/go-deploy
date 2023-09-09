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

func waitForDeploymentCreated(t *testing.T, id string, callback func(*body.DeploymentRead) bool) {
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
