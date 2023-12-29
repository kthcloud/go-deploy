package github

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	"go-deploy/test/e2e"
	"net/http"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	e2e.Setup()
	code := m.Run()
	e2e.Shutdown()
	os.Exit(code)
}

func TestFetchRepositories(t *testing.T) {
	t.Parallel()

	// if you want to test this, you need to set the code below.
	// this code should be given to you by authorizing towards GitHub
	code := ""

	//goland:noinspection GoBoolExpressions
	if code == "" {
		t.Skip("skipping test; code not set")
	}

	resp := e2e.DoGetRequest(t, "/github/repositories?"+code)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var readBody body.GitHubRepositoriesRead
	err := e2e.ReadResponseBody(t, resp, &readBody)
	assert.NoError(t, err, "repositories were not fetched")

	assert.NotEmpty(t, readBody.AccessToken, "access token was empty")
	for _, repository := range readBody.Repositories {
		assert.NotEmpty(t, repository.ID, "repository id was empty")
		assert.NotEmpty(t, repository.Name, "repository name was empty")
	}
}
