package harbor

import (
	"github.com/kthcloud/go-deploy/test"
	"github.com/kthcloud/go-deploy/test/acc"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateWebhook(t *testing.T) {
	c, _ := withContext(t)
	withDefaultWebhook(t, c)
}

func TestUpdateWebhook(t *testing.T) {
	c, _ := withContext(t)
	w := withDefaultWebhook(t, c)

	w.Name = acc.GenName()
	w.Target = "https://some-other-url.com"
	w.Token = acc.GenName()

	wUpdated, err := c.UpdateWebhook(w)
	test.NoError(t, err, "failed to update webhook")

	assert.Equal(t, w.Name, wUpdated.Name, "webhook name does not match")
	assert.Equal(t, w.Target, wUpdated.Target, "webhook target does not match")
	assert.Equal(t, w.Token, wUpdated.Token, "webhook token does not match")
}
