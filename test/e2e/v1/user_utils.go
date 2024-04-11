package v1

import (
	"go-deploy/dto/v1/body"
	"go-deploy/test/e2e"
	"testing"
)

const (
	UserPath  = "/v1/users/"
	UsersPath = "/v1/users"
)

func GetUser(t *testing.T, id string) body.UserRead {
	resp := e2e.DoGetRequest(t, UserPath+id)
	return e2e.Parse[body.UserRead](t, resp)
}

func ListUsers(t *testing.T, query string) []body.UserRead {
	resp := e2e.DoGetRequest(t, UsersPath+query)
	return e2e.Parse[[]body.UserRead](t, resp)
}

func ListUsersDiscovery(t *testing.T, query string) []body.UserReadDiscovery {
	if query == "" {
		query = "?"
	}

	resp := e2e.DoGetRequest(t, UsersPath+query+"&discovery=true")
	return e2e.Parse[[]body.UserReadDiscovery](t, resp)
}
