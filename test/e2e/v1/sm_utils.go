package v1

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v1/body"
	"go-deploy/test/e2e"
	"net/http"
	"testing"
)

const (
	SmPath  = "/v1/storageManagers/"
	SmsPath = "/v1/storageManagers"
)

func GetSM(t *testing.T, id string) body.SmRead {
	resp := e2e.DoGetRequest(t, SmPath+id)
	return e2e.Parse[body.SmRead](t, resp)
}

func ListSMs(t *testing.T, query string) []body.SmRead {
	resp := e2e.DoGetRequest(t, SmsPath+query)
	return e2e.Parse[[]body.SmRead](t, resp)
}

func WaitForSmRunning(t *testing.T, id string, callback func(read *body.SmRead) bool) {
	e2e.FetchUntil(t, SmPath+id, func(resp *http.Response) bool {
		var smRead body.SmRead
		err := e2e.ReadResponseBody(t, resp, &smRead)
		assert.NoError(t, err, "storage manager was not fetched")

		if callback == nil {
			return true
		}

		return callback(&smRead)
	})
}
