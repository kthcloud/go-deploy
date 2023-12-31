package k8s

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/test"
	"go-deploy/test/acc"
	"testing"
)

func TestCreateIngress(t *testing.T) {
	t.Parallel()

	c, _ := withContext(t)
	withDefaultIngress(t, c, withDefaultService(t, c))
}

func TestUpdateIngress(t *testing.T) {
	t.Parallel()

	c, _ := withContext(t)
	i := withDefaultIngress(t, c, withDefaultService(t, c))

	i.Hosts = []string{acc.GenName() + ".another.com"}

	iUpdated, err := c.UpdateIngress(i)
	test.NoError(t, err, "failed to update ingress")

	test.EqualOrEmpty(t, i.Hosts, iUpdated.Hosts, "hosts do not match")
}

func TestUpdateIngressService(t *testing.T) {
	t.Parallel()

	c, _ := withContext(t)
	i := withDefaultIngress(t, c, withDefaultService(t, c))

	i.ServiceName = acc.GenName()
	i.ServicePort = 12345

	iUpdated, err := c.UpdateIngress(i)
	test.NoError(t, err, "failed to update ingress")

	assert.Equal(t, i.ServiceName, iUpdated.ServiceName, "service name does not match")
	assert.Equal(t, i.ServicePort, iUpdated.ServicePort, "service port does not match")
}
