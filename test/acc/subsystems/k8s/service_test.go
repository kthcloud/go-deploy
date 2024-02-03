package k8s

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/test"
	"testing"
)

func TestCreateService(t *testing.T) {
	t.Parallel()

	c, _ := withContext(t)
	withDefaultService(t, c)
}

func TestUpdateService(t *testing.T) {
	t.Parallel()

	c, _ := withContext(t)
	s := withDefaultService(t, c)

	s.Ports[0].Port = 12345
	s.Ports[0].TargetPort = 12345

	sUpdated, err := c.UpdateService(s)
	test.NoError(t, err, "failed to update service")

	assert.Equal(t, s.Ports[0].Port, sUpdated.Ports[0].Port, "service port does not match")
	assert.Equal(t, s.Ports[0].TargetPort, sUpdated.Ports[0].TargetPort, "service target port does not match")
}
