package v2

import (
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/test/e2e"
	"testing"
)

const (
	ZonePath  = "/v2/zones/"
	ZonesPath = "/v2/zones"
)

func GetZone(t *testing.T, id string) body.ZoneRead {
	resp := e2e.DoGetRequest(t, ZonePath+id)
	return e2e.MustParse[body.ZoneRead](t, resp)
}

func ListZones(t *testing.T, query string) []body.ZoneRead {
	resp := e2e.DoGetRequest(t, ZonesPath+query)
	return e2e.MustParse[[]body.ZoneRead](t, resp)
}
