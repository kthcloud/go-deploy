package e2e

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	"go-deploy/pkg/app/status_codes"
	"net/http"
	"testing"
)

func GetJob(t *testing.T, id string, userID ...string) body.JobRead {
	resp := DoGetRequest(t, "/jobs/"+id, userID...)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "job was not fetched")

	var jobRead body.JobRead
	err := ReadResponseBody(t, resp, &jobRead)
	assert.NoError(t, err, "job was not fetched")

	return jobRead
}

func ListJobs(t *testing.T, query string, userID ...string) []body.JobRead {
	resp := DoGetRequest(t, "/jobs"+query, userID...)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "jobs were not fetched")

	var jobs []body.JobRead
	err := ReadResponseBody(t, resp, &jobs)
	assert.NoError(t, err, "jobs were not fetched")

	return jobs
}

func UpdateJob(t *testing.T, id string, requestBody body.JobUpdate, userID ...string) body.JobRead {
	resp := DoPostRequest(t, "/jobs/"+id, requestBody, userID...)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "job was not updated")

	var job body.JobRead
	err := ReadResponseBody(t, resp, &job)
	assert.NoError(t, err, "job was not updated")

	if requestBody.Status != nil {
		assert.Equal(t, *requestBody.Status, job.Status, "job status was not updated")
	}

	return job
}

func WaitForJobFinished(t *testing.T, id string, callback func(*body.JobRead) bool) {
	fetchUntil(t, "/jobs/"+id, func(resp *http.Response) bool {
		jobRead := Parse[body.JobRead](t, resp)

		if jobRead.Status == status_codes.GetMsg(status_codes.JobFinished) || jobRead.Status == status_codes.GetMsg(status_codes.JobTerminated) {
			if callback == nil || callback(&jobRead) {
				return true
			}
		}

		return false
	})
}
