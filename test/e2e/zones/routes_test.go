package zones

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v1/body"
	"go-deploy/test/e2e"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	e2e.Setup()
	code := m.Run()
	e2e.Shutdown()
	os.Exit(code)
}

func TestList(t *testing.T) {
	resp := e2e.DoGetRequest(t, "/zones")
	assert.Equal(t, 200, resp.StatusCode)

	var zones []body.ZoneRead
	err := e2e.ReadResponseBody(t, resp, &zones)
	assert.NoError(t, err, "zones were not fetched")

	assert.NotEmpty(t, zones, "zones were not fetched. it should have at least one zone")

	for _, zone := range zones {
		assert.NotEmpty(t, zone.Name, "zone id was empty")
		assert.NotEmpty(t, zone.Type, "zone type was empty")

		if zone.Type != "vm" && zone.Type != "deployment" {
			assert.Fail(t, "zone type was invalid. it should be either vm or deployment")
		}
	}
}
