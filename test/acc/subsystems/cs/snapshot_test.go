package cs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateSnapshot(t *testing.T) {
	t.Parallel()

	withDefaultSnapshot(t, withDefaultVM(t, withCsServiceOfferingSmall(t)))
}

func TestRestoreSnapshot(t *testing.T) {
	t.Parallel()

	client := withClient(t)
	vm := withDefaultVM(t, withCsServiceOfferingSmall(t))
	snapshot := withDefaultSnapshot(t, vm)

	err := client.ApplySnapshot(snapshot)
	assert.NoError(t, err, "failed to restore snapshot")
}
