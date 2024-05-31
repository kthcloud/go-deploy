package v2

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v2/body"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/test/e2e"
	"net/http"
	"testing"
)

const (
	JobPath  = "/v2/jobs/"
	JobsPath = "/v2/jobs"
)

func GetJob(t *testing.T, id string, user ...string) body.JobRead {
	resp := e2e.DoGetRequest(t, JobPath+id, user...)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "job was not fetched")

	var jobRead body.JobRead
	err := e2e.ReadResponseBody(t, resp, &jobRead)
	assert.NoError(t, err, "job was not fetched")

	return jobRead
}

func ListJobs(t *testing.T, query string, user ...string) []body.JobRead {
	resp := e2e.DoGetRequest(t, JobsPath+query, user...)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "jobs were not fetched")

	var jobs []body.JobRead
	err := e2e.ReadResponseBody(t, resp, &jobs)
	assert.NoError(t, err, "jobs were not fetched")

	return jobs
}

func UpdateJob(t *testing.T, id string, requestBody body.JobUpdate, user ...string) body.JobRead {
	resp := e2e.DoPostRequest(t, JobPath+id, requestBody, user...)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "job was not updated")

	var job body.JobRead
	err := e2e.ReadResponseBody(t, resp, &job)
	assert.NoError(t, err, "job was not updated")

	if requestBody.Status != nil {
		assert.Equal(t, *requestBody.Status, job.Status, "job status was not updated")
	}

	return job
}

func WaitForJobFinished(t *testing.T, id string, callback func(*body.JobRead) bool) {
	e2e.FetchUntil(t, JobPath+id, func(resp *http.Response) bool {
		jobRead := e2e.MustParse[body.JobRead](t, resp)

		if jobRead.Status == status_codes.GetMsg(status_codes.JobFinished) || jobRead.Status == status_codes.GetMsg(status_codes.JobTerminated) {
			if callback == nil || callback(&jobRead) {
				return true
			}
		}

		return false
	})
}
