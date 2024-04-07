package v1

import (
	"go-deploy/dto/v1/body"
	"go-deploy/test/e2e"
	"testing"
)

func GetSM(t *testing.T, id string) body.SmRead {
	resp := e2e.DoGetRequest(t, "/storageManagers/"+id)
	return e2e.Parse[body.SmRead](t, resp)
}

func ListSMs(t *testing.T, query string) []body.SmRead {
	resp := e2e.DoGetRequest(t, "/storageManagers"+query)
	return e2e.Parse[[]body.SmRead](t, resp)
}
