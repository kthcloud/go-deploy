package gpu_leases

import (
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/test/e2e"
	v2 "github.com/kthcloud/go-deploy/test/e2e/v2"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if e2e.VmTestsEnabled {
		e2e.Setup()
		code := m.Run()
		e2e.Shutdown()
		os.Exit(code)
	}
}

func TestGetGpuLease(t *testing.T) {
	t.Parallel()

	queries := []string{
		"?page=1&pageSize=10",
		"?userId=" + model.TestPowerUserID + "&page=1&pageSize=3",
		"?userId=" + model.TestPowerUserID + "&page=1&pageSize=3",
	}

	for _, query := range queries {
		v2.ListGpuLeases(t, query)
	}
}

func TestCreateGpuLease(t *testing.T) {
	// Can't be run in parallel because a user may only have one GPU lease at the same time at the moment
	// t.Parallel()

	gpuGroup := v2.GetAnyAvailableGpuGroup(t)
	if gpuGroup == nil {
		t.Skip("No available GPU group")
		return
	}

	v2.WithDefaultGpuLease(t, gpuGroup.ID)
}

func TestActivateGpuLease(t *testing.T) {
	// Can't be run in parallel because a user may only have one GPU lease at the same time at the moment
	// t.Parallel()

	// We don't want to hog a GPU group only for tests
	// so we will only test if there is a GPU group specified manually
	gpuGroupID := ""
	if gpuGroupID == "" {
		t.Skip("No GPU group specified for activating GPU lease")
		return
	}

	gpuGroup := v2.GetGpuGroup(t, gpuGroupID)
	gpuLease := v2.WithDefaultGpuLease(t, gpuGroup.ID)
	vm := v2.WithDefaultVM(t)

	requestBody := body.GpuLeaseUpdate{
		VmID: &vm.ID,
	}

	v2.UpdateGpuLease(t, gpuLease.ID, requestBody)
}
