package v1

import (
	"go-deploy/dto/v1/body"
	"go-deploy/test/e2e"
	"testing"
)

const (
	ZonePath  = "/v1/zones/"
	ZonesPath = "/v1/zones"
)

func GetZone(t *testing.T, id string) body.ZoneRead {
	resp := e2e.DoGetRequest(t, ZonePath+id)
	return e2e.Parse[body.ZoneRead](t, resp)
}

func ListZones(t *testing.T, query string) []body.ZoneRead {
	resp := e2e.DoGetRequest(t, ZonesPath+query)
	return e2e.Parse[[]body.ZoneRead](t, resp)
}
