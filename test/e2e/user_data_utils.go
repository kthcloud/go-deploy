package e2e

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v1/body"
	"net/http"
	"testing"
)

func GetUserData(t *testing.T, id string, userID ...string) body.UserDataRead {
	resp := DoGetRequest(t, "/userData/"+id, userID...)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "user data was not fetched")

	var userDataRead body.UserDataRead
	err := ReadResponseBody(t, resp, &userDataRead)
	assert.NoError(t, err, "user data was not fetched")

	return userDataRead
}

func ListUserData(t *testing.T, query string, userID ...string) []body.UserDataRead {
	resp := DoGetRequest(t, "/userData"+query, userID...)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "user data was not fetched")

	var userData []body.UserDataRead
	err := ReadResponseBody(t, resp, &userData)
	assert.NoError(t, err, "user data was not fetched")

	return userData
}

func UpdateUserData(t *testing.T, id string, userDataUpdate body.UserDataUpdate, userID ...string) body.UserDataRead {
	resp := DoPostRequest(t, "/userData/"+id, userDataUpdate, userID...)
	var userDataRead body.UserDataRead
	err := ReadResponseBody(t, resp, &userDataRead)
	assert.NoError(t, err, "user data was not updated")

	assert.Equal(t, userDataUpdate.Data, userDataRead.Data, "invalid user data name")

	return userDataRead
}

func DeleteUserData(t *testing.T, id string, userID ...string) {
	resp := DoDeleteRequest(t, "/userData/"+id, userID...)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode, "user data was not deleted")
}

func WithUserData(t *testing.T, userDataCreate body.UserDataCreate, userID ...string) body.UserDataRead {
	resp := DoPostRequest(t, "/userData", userDataCreate, userID...)
	userData := Parse[body.UserDataRead](t, resp)

	assert.Equal(t, userDataCreate.Data, userData.Data, "invalid user data name")

	t.Cleanup(func() {
		DeleteUserData(t, userData.ID, userID...)
	})

	return userData
}
