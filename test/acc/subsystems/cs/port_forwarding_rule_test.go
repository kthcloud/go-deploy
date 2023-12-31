package cs

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/test"
	"testing"
)

func TestCreatePortForwardingRule(t *testing.T) {
	t.Parallel()

	withDefaultPFR(t, withDefaultVM(t, withCsServiceOfferingSmall(t)))
}

func TestUpdatePortForwardingRule(t *testing.T) {
	t.Parallel()

	client := withClient(t)
	vm := withDefaultVM(t, withCsServiceOfferingSmall(t))
	pfr := withDefaultPFR(t, vm)

	pfr.PrivatePort = pfr.PrivatePort + 1

	pfrUpdated, err := client.UpdatePortForwardingRule(pfr)
	test.NoError(t, err, "failed to update port forwarding rule")

	assert.Equal(t, pfr.PrivatePort, pfrUpdated.PrivatePort, "port forwarding rule is not updated")
}
