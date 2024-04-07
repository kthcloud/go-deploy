package v1

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v1/body"
	"go-deploy/test/e2e"
	"net/http"
	"testing"
)

func GetUserData(t *testing.T, id string, userID ...string) body.UserDataRead {
	resp := e2e.DoGetRequest(t, "/userData/"+id, userID...)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "user data was not fetched")

	var userDataRead body.UserDataRead
	err := e2e.ReadResponseBody(t, resp, &userDataRead)
	assert.NoError(t, err, "user data was not fetched")

	return userDataRead
}

func ListUserData(t *testing.T, query string, userID ...string) []body.UserDataRead {
	resp := e2e.DoGetRequest(t, "/userData"+query, userID...)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "user data was not fetched")

	var userData []body.UserDataRead
	err := e2e.ReadResponseBody(t, resp, &userData)
	assert.NoError(t, err, "user data was not fetched")

	return userData
}

func UpdateUserData(t *testing.T, id string, userDataUpdate body.UserDataUpdate, userID ...string) body.UserDataRead {
	resp := e2e.DoPostRequest(t, "/userData/"+id, userDataUpdate, userID...)
	var userDataRead body.UserDataRead
	err := e2e.ReadResponseBody(t, resp, &userDataRead)
	assert.NoError(t, err, "user data was not updated")

	assert.Equal(t, userDataUpdate.Data, userDataRead.Data, "invalid user data name")

	return userDataRead
}

func DeleteUserData(t *testing.T, id string, userID ...string) {
	resp := e2e.DoDeleteRequest(t, "/userData/"+id, userID...)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode, "user data was not deleted")
}

func WithUserData(t *testing.T, userDataCreate body.UserDataCreate, userID ...string) body.UserDataRead {
	resp := e2e.DoPostRequest(t, "/userData", userDataCreate, userID...)
	userData := e2e.Parse[body.UserDataRead](t, resp)

	assert.Equal(t, userDataCreate.Data, userData.Data, "invalid user data name")

	t.Cleanup(func() {
		DeleteUserData(t, userData.ID, userID...)
	})

	return userData
}
