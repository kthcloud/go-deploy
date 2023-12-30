package harbor

import (
	"testing"
)

func TestCreateRepository(t *testing.T) {
	c, _ := withContext(t)
	withDefaultRepository(t, c)
}

func TestUpdateRepository(t *testing.T) {
	t.Skip("no fields can be updated right now")
}
