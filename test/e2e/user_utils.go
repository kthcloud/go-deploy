package e2e

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	"net/http"
	"testing"
)

func GetUser(t *testing.T, id string) body.UserRead {
	resp := DoGetRequest(t, "/users/"+id)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "user was not fetched")

	var userRead body.UserRead
	err := ReadResponseBody(t, resp, &userRead)
	assert.NoError(t, err, "user was not fetched")

	return userRead
}

func ListUsers(t *testing.T, query string) []body.UserRead {
	resp := DoGetRequest(t, "/users"+query)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "users were not fetched")

	var users []body.UserRead
	err := ReadResponseBody(t, resp, &users)
	assert.NoError(t, err, "users were not fetched")

	return users
}
