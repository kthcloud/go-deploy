package users

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	"go-deploy/test/e2e"
	"net/http"
	"testing"
)

func TestFetchUsers(t *testing.T) {
	resp := e2e.DoGetRequest(t, "/users")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var users []body.UserRead
	err := e2e.ReadResponseBody(t, resp, &users)
	assert.NoError(t, err, "users were not fetched")

	assert.NotEmpty(t, users, "users were not fetched. it should have at least one user (test user)")

	for _, user := range users {
		assert.NotEmpty(t, user.ID, "user id was empty")
	}
}

func TestFetchUser(t *testing.T) {
	resp := e2e.DoGetRequest(t, "/users/test")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var userRead body.UserRead
	err := e2e.ReadResponseBody(t, resp, &userRead)
	assert.NoError(t, err, "user was not fetched")

	assert.Equal(t, e2e.TestUserID, userRead.ID, "invalid user id")
}
