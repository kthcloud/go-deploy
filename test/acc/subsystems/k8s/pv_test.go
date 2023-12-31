package k8s

import "testing"

func TestCreatePV(t *testing.T) {
	t.Parallel()

	c, _ := withContext(t)
	withDefaultPV(t, c)
}

func TestUpdatePV(t *testing.T) {
	t.Skip("no fields can be updated right now")
}
