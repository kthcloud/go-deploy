package k8s

import (
	"github.com/stretchr/testify/assert"
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

	s.Port = 12345
	s.TargetPort = 12345

	sUpdated, err := c.UpdateService(s)
	assert.NoError(t, err, "failed to update service")

	assert.Equal(t, s.Port, sUpdated.Port, "service port does not match")
	assert.Equal(t, s.TargetPort, sUpdated.TargetPort, "service target port does not match")
}
