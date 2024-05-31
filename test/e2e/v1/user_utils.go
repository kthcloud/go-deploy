package v1

import (
	"github.com/stretchr/testify/assert"
	body2 "go-deploy/dto/v2/body"
	"go-deploy/test"
	"go-deploy/test/e2e"
	"testing"
)

const (
	UserPath  = "/v1/users/"
	UsersPath = "/v1/users"
)

func GetUser(t *testing.T, id string, user ...string) body2.UserRead {
	resp := e2e.DoGetRequest(t, UserPath+id, user...)
	return e2e.MustParse[body2.UserRead](t, resp)
}

func ListUsers(t *testing.T, query string) []body2.UserRead {
	resp := e2e.DoGetRequest(t, UsersPath+query)
	return e2e.MustParse[[]body2.UserRead](t, resp)
}

func ListUsersDiscovery(t *testing.T, query string) []body2.UserReadDiscovery {
	if query == "" {
		query = "?"
	}

	resp := e2e.DoGetRequest(t, UsersPath+query+"&discovery=true")
	return e2e.MustParse[[]body2.UserReadDiscovery](t, resp)
}

func UpdateUser(t *testing.T, id string, update body2.UserUpdate) body2.UserRead {
	resp := e2e.DoPostRequest(t, UserPath+id, update)
	userRead := e2e.MustParse[body2.UserRead](t, resp)

	if update.PublicKeys != nil {
		foundAll := true
		for _, key := range userRead.PublicKeys {
			found := false
			for _, updateKey := range *update.PublicKeys {
				if key.Name == updateKey.Name && key.Key == updateKey.Key {
					found = true
					break
				}
			}

			if !found {
				foundAll = false
				break
			}
		}
		assert.True(t, foundAll, "public keys were not updated")
	}

	if update.UserData != nil {
		foundAll := true
		for _, data := range userRead.UserData {
			found := false
			for _, updateData := range *update.UserData {
				if data.Key == updateData.Key && data.Value == updateData.Value {
					found = true
					break
				}
			}

			if !found {
				foundAll = false
				break
			}
		}
		assert.True(t, foundAll, "user data were not updated")
	}

	return userRead
}

func CreateApiKey(t *testing.T, userID string, apiKeyCreate body2.ApiKeyCreate) {
	resp := e2e.DoPostRequest(t, UserPath+userID+"/apiKeys", apiKeyCreate)
	apiKeyCreated := e2e.MustParse[body2.ApiKeyCreated](t, resp)

	assert.NotEmpty(t, apiKeyCreated.Key)
	assert.Equal(t, apiKeyCreate.Name, apiKeyCreated.Name)
	test.TimeNotZero(t, apiKeyCreated.CreatedAt)
	test.TimeEq(t, apiKeyCreated.ExpiresAt, apiKeyCreated.ExpiresAt)
}
