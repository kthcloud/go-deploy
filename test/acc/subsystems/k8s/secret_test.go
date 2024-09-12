package k8s

import (
	"github.com/stretchr/testify/assert"
	"github.com/kthcloud/go-deploy/test"
	"github.com/kthcloud/go-deploy/test/acc"
	"testing"
)

func TestCreateSecret(t *testing.T) {
	t.Parallel()

	c, _ := withContext(t)
	withDefaultSecret(t, c)
}

func TestUpdateSecret(t *testing.T) {
	t.Parallel()

	c, _ := withContext(t)
	s := withDefaultSecret(t, c)

	key := acc.GenName()
	s.Data[key] = []byte(acc.GenName())

	sUpdated, err := c.UpdateSecret(s)
	test.NoError(t, err, "failed to update secret")

	_, ok := sUpdated.Data[key]
	assert.True(t, ok, "secret data does not match")
}
