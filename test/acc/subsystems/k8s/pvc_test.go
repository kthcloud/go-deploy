//go:build acc

package k8s

import "testing"

func TestCreatePVC(t *testing.T) {
	t.Parallel()

	c, _ := withContext(t)
	withDefaultPVC(t, c, withDefaultPV(t, c))
}

func TestUpdatePVC(t *testing.T) {
	t.Skip("no fields can be updated right now")
}
