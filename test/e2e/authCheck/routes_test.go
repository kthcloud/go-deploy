package authCheck

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/service"
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

func TestFetchAuthCheck(t *testing.T) {
	resp := e2e.DoGetRequest(t, "/authCheck")
	assert.Equal(t, 200, resp.StatusCode)

	var authInfo service.AuthInfo
	err := e2e.ReadResponseBody(t, resp, &authInfo)
	assert.NoError(t, err, "auth info was not fetched")

	assert.NotEmpty(t, authInfo.UserID, "user id was empty")
	assert.NotEmpty(t, authInfo.JwtToken, "jwt token was empty")
}
