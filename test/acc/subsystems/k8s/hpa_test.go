package k8s

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/test"
	"testing"
)

func TestCreateHPA(t *testing.T) {
	t.Parallel()

	c, _ := withContext(t)
	withDefaultHPA(t, c, withDefaultDeployment(t, c))
}

func TestUpdateHPA(t *testing.T) {
	t.Parallel()

	c, _ := withContext(t)
	h := withDefaultHPA(t, c, withDefaultDeployment(t, c))

	h.MinReplicas = 2
	h.MaxReplicas = 4
	h.CpuAverageUtilization = 50
	h.MemoryAverageUtilization = 50

	hUpdated, err := c.UpdateHPA(h)
	test.NoError(t, err, "failed to update hpa")

	assert.Equal(t, h.MinReplicas, hUpdated.MinReplicas, "hpa min replicas does not match")
	assert.Equal(t, h.MaxReplicas, hUpdated.MaxReplicas, "hpa max replicas does not match")
	assert.Equal(t, h.CpuAverageUtilization, hUpdated.CpuAverageUtilization, "hpa cpu average utilization does not match")
	assert.Equal(t, h.MemoryAverageUtilization, hUpdated.MemoryAverageUtilization, "hpa memory average utilization does not match")
}
