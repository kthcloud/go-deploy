package v2

import (
	"go-deploy/dto/v2/body"
	"go-deploy/test/e2e"
	"net/http"
	"testing"
)

const (
	SmPath  = "/v2/storageManagers/"
	SmsPath = "/v2/storageManagers"
)

func GetSM(t *testing.T, id string) body.SmRead {
	resp := e2e.DoGetRequest(t, SmPath+id)
	return e2e.MustParse[body.SmRead](t, resp)
}

func ListSMs(t *testing.T, query string) []body.SmRead {
	resp := e2e.DoGetRequest(t, SmsPath+query)
	return e2e.MustParse[[]body.SmRead](t, resp)
}

func WaitForSmRunning(t *testing.T, id string, callback func(read *body.SmRead) bool) {
	e2e.FetchUntil(t, SmPath+id, func(resp *http.Response) bool {
		smRead := e2e.MustParse[body.SmRead](t, resp)
		if callback == nil {
			return true
		}

		return callback(&smRead)
	})
}
