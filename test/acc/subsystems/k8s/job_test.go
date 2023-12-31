package k8s

import "testing"

func TestCreateJob(t *testing.T) {
	t.Parallel()

	c, _ := withContext(t)
	withDefaultJob(t, c)
}

func TestUpdateJob(t *testing.T) {
	t.Skip("no fields can be updated right now")
}
