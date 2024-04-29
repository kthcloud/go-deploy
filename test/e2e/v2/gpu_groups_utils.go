package v2

import (
	"go-deploy/dto/v2/body"
	"go-deploy/test/e2e"
	"testing"
)

const (
	GpuGroupPath  = "/v2/gpuGroups/"
	GpuGroupsPath = "/v2/gpuGroups"
)

func GetGpuGroup(t *testing.T, id string, userID ...string) body.GpuGroupRead {
	resp := e2e.DoGetRequest(t, GpuGroupPath+id, userID...)
	return e2e.MustParse[body.GpuGroupRead](t, resp)
}

func GetAnyAvailableGpuGroup(t *testing.T, userID ...string) *body.GpuGroupRead {
	gpuGroups := ListGpuGroups(t, "?available=true", userID...)
	if len(gpuGroups) == 0 {
		return nil
	}

	return &gpuGroups[0]
}

func ListGpuGroups(t *testing.T, query string, userID ...string) []body.GpuGroupRead {
	resp := e2e.DoGetRequest(t, GpuGroupsPath+query, userID...)
	return e2e.MustParse[[]body.GpuGroupRead](t, resp)
}
