package cs

import (
	"go-deploy/test"
	"testing"
)

func TestCreateSnapshot(t *testing.T) {
	t.Skip("CloudStack is too unpredictable to run this test")

	t.Parallel()

	withDefaultSnapshot(t, withDefaultVM(t))
}

func TestRestoreSnapshot(t *testing.T) {
	t.Skip("CloudStack is too unpredictable to run this test")

	t.Parallel()

	client := withClient(t)
	vm := withDefaultVM(t)
	snapshot := withDefaultSnapshot(t, vm)

	err := client.ApplySnapshot(snapshot)
	test.NoError(t, err, "failed to restore snapshot")
}
