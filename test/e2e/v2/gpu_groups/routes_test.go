package gpu_groups

import (
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

func TestGetGpuGroup(t *testing.T) {
	t.Parallel()

	v2.GetAnyAvailableGpuGroup(t)
}

func TestListGpuGroups(t *testing.T) {
	t.Parallel()

	queries := []string{
		"?page=1&pageSize=10",
	}

	for _, query := range queries {
		v2.ListGpuGroups(t, query)
	}
}
