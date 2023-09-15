package cs

import "testing"

func TestCreatePortForwardingRule(t *testing.T) {

	vm := withVM(t, withCsServiceOfferingType1(t))
	pfr := withPortForwardingRule(t, vm)
	defer func() {
		cleanUpPortForwardingRule(t, pfr.ID)
		cleanUpVM(t, vm.ID)
	}()
}
