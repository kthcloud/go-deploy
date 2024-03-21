package e2e

import (
	"go-deploy/dto/v1/body"
	"testing"
)

func GetSM(t *testing.T, id string) body.SmRead {
	resp := DoGetRequest(t, "/storageManagers/"+id)
	return Parse[body.SmRead](t, resp)
}

func ListSMs(t *testing.T, query string) []body.SmRead {
	resp := DoGetRequest(t, "/storageManagers"+query)
	return Parse[[]body.SmRead](t, resp)
}
