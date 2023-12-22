package e2e

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	status_codes2 "go-deploy/pkg/app/status_codes"
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

func WaitForJobFinished(t *testing.T, id string, callback func(*body.JobRead) bool) {
	fetchUntil(t, "/jobs/"+id, func(resp *http.Response) bool {
		var jobRead body.JobRead
		err := ReadResponseBody(t, resp, &jobRead)
		assert.NoError(t, err, "job was not fetched")

		if jobRead.Status == status_codes2.GetMsg(status_codes2.JobFinished) || jobRead.Status == status_codes2.GetMsg(status_codes2.JobTerminated) {
			if callback == nil || callback(&jobRead) {
				return true
			}
		}

		return false
	})
}
