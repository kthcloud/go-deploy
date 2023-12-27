package users

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

func TestGet(t *testing.T) {
	resp := e2e.DoGetRequest(t, "/users/"+e2e.AdminUserID)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var userRead body.UserRead
	err := e2e.ReadResponseBody(t, resp, &userRead)
	assert.NoError(t, err, "user was not fetched")

	assert.Equal(t, e2e.AdminUserID, userRead.ID, "invalid user id")
}

func TestList(t *testing.T) {
	queries := []string{
		// all
		"?all=true",
		// search
		"?search=tester",
	}

	for _, query := range queries {
		users := e2e.ListUsers(t, query)
		assert.NotEmpty(t, users, "users were not fetched. it should have at least one user")
	}
}

func TestDiscover(t *testing.T) {
	// Since Discover does not return yourself, we need to ensure that some other user exists in the database
	// This is easiest done by doing any GET request fora user
	resp := e2e.DoGetRequest(t, "/users/"+e2e.DefaultUserID, e2e.DefaultUserID)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp = e2e.DoGetRequest(t, "/users?discover=true&search=tester")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var users []body.UserRead
	err := e2e.ReadResponseBody(t, resp, &users)
	assert.NoError(t, err, "users were not fetched")

	assert.NotEmpty(t, users, "users were not fetched. it should have at least one user")
}
