package k8s

import "testing"

func TestCreateNamespace(t *testing.T) {
	t.Parallel()

	withNamespace(t)
}

func TestUpdateNamespace(t *testing.T) {
	t.Skip("no fields can be updated right now")
}
