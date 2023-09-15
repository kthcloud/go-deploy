package cs

import (
	"testing"
)

func TestCreateServiceOffering(t *testing.T) {
	so := withCsServiceOfferingType1(t)
	defer cleanUpServiceOffering(t, so.ID)
}
