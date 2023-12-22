package e2e

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	"net/http"
	"testing"
)

func GetSM(t *testing.T, id string) body.SmRead {
	resp := DoGetRequest(t, "/storageManagers/"+id)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "sm was not fetched")

	var smRead body.SmRead
	err := ReadResponseBody(t, resp, &smRead)
	assert.NoError(t, err, "storage manager was not fetched")

	return smRead
}
