package v2

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v2/body"
	"go-deploy/test/e2e"
	v1 "go-deploy/test/e2e/v1"
	"testing"
)

const (
	GpuLeasePath  = "/v2/gpuLeases/"
	GpuLeasesPath = "/v2/gpuLeases"
)

func GetGpuLease(t *testing.T, id string, userID ...string) body.GpuLeaseRead {
	resp := e2e.DoGetRequest(t, GpuLeasePath+id, userID...)
	return e2e.MustParse[body.GpuLeaseRead](t, resp)
}

func ListGpuLeases(t *testing.T, query string, userID ...string) []body.GpuLeaseRead {
	resp := e2e.DoGetRequest(t, GpuLeasesPath+query, userID...)
	return e2e.MustParse[[]body.GpuLeaseRead](t, resp)
}

func UpdateGpuLease(t *testing.T, id string, requestBody body.GpuLeaseUpdate, userID ...string) body.GpuLeaseRead {
	resp := e2e.DoPostRequest(t, GpuLeasePath+id, requestBody, userID...)
	gpuLeaseUpdated := e2e.MustParse[body.GpuLeaseUpdated](t, resp)

	v1.WaitForJobFinished(t, gpuLeaseUpdated.JobID, nil)

	gpuLease := GetGpuLease(t, id, userID...)
	assert.Equal(t, requestBody.VmID, gpuLease.VmID)

	return GetGpuLease(t, id, userID...)
}

func WithDefaultGpuLease(t *testing.T, gpuGroupID string, userID ...string) body.GpuLeaseRead {
	requestBody := body.GpuLeaseCreate{
		GpuGroupID:   gpuGroupID,
		LeaseForever: false,
	}

	return WithGpuLease(t, requestBody, userID...)
}

func WithGpuLease(t *testing.T, requestBody body.GpuLeaseCreate, userID ...string) body.GpuLeaseRead {
	resp := e2e.DoPostRequest(t, GpuLeasesPath, requestBody, userID...)
	gpuLeaseCreated := e2e.MustParse[body.GpuLeaseCreated](t, resp)

	t.Cleanup(func() { cleanUpGpuLease(t, gpuLeaseCreated.ID) })

	v1.WaitForJobFinished(t, gpuLeaseCreated.JobID, nil)

	gpuLeaseRead := GetGpuLease(t, gpuLeaseCreated.ID, userID...)

	assert.NotEmpty(t, gpuLeaseRead.ID)
	assert.Equal(t, requestBody.GpuGroupID, gpuLeaseRead.GpuGroupID)
	if requestBody.LeaseForever {
		// Forever leases should at least be more than 1000 hours
		assert.Greater(t, gpuLeaseRead.LeaseDuration, 1000)
	}

	return gpuLeaseRead
}

func cleanUpGpuLease(t *testing.T, id string) {
	resp := e2e.DoDeleteRequest(t, GpuLeasePath+id)
	v1.WaitForJobFinished(t, e2e.MustParse[body.GpuLeaseDeleted](t, resp).JobID, nil)
}
