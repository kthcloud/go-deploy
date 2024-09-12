package harbor

import (
	"github.com/stretchr/testify/assert"
	"github.com/kthcloud/go-deploy/test"
	"testing"
)

func TestCreateProject(t *testing.T) {
	withDefaultProject(t)
}

func TestUpdateProject(t *testing.T) {
	c, p := withContext(t)

	p.Public = !p.Public

	pUpdated, err := c.UpdateProject(p)
	test.NoError(t, err, "failed to update project")

	assert.Equal(t, p.Public, pUpdated.Public, "project public does not match")
}
